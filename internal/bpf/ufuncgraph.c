// +build ignore

#include "common.h"
#include "bpf_tracing.h"

#define MAX_STACK_LAYERS 5000
#define MAX_GOROUTINES 1000

char __license[] SEC("license") = "Dual MIT/GPL";

struct event {
    __u64 stack_id;
    __u64 caller_ip;
    __u64 ip;
    __u64 time_ns;
    __u32 stack_depth;
    __u16 hook_point; // 0 entry; 1 exit;
    __u16 errno; // 0 no error; 1 root bp; 2 stackoverflow;
    __u8 args[104];
};

// force emitting struct event into the ELF.
const struct event *unused __attribute__((unused));

struct stackwalk {
    __u64 depth;
    __u64 root_bp;
    __u64 stack_id;
};

// event_queue is for events commit
struct bpf_map_def SEC("maps") event_queue = {
    .type = BPF_MAP_TYPE_QUEUE,
    .key_size = 0,
    .value_size = sizeof(struct event),
    .max_entries = 1000000,
};

// goids is for stack_id generation
struct bpf_map_def SEC("maps") goids = {
    .type = BPF_MAP_TYPE_PERCPU_ARRAY,
    .key_size = sizeof(__u32),
    .value_size = sizeof(__u64),
    .max_entries = 1,
};

// bp_to_event is for looking up the caller event
struct bpf_map_def SEC("maps") bp_to_event = {
    .type = BPF_MAP_TYPE_LRU_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct event),
    .max_entries = MAX_GOROUTINES,
};

void static backtrace(__u64 bp, struct stackwalk *walk) {
    for (walk->depth = 0; walk->depth < MAX_STACK_LAYERS; walk->depth++) {
        if (bpf_probe_read_user(&walk->stack_id, sizeof(walk->stack_id), (void*)bp) < 0) {
            walk->stack_id = bp;
            return;
        }
        walk->root_bp = bp;
        bp = walk->stack_id;
    }
    walk->depth = 0xffffffffffffffff;
    return;
}

__u64 static new_stack_id() {
    __u32 key = 0;
    __u32 *stack_id = bpf_map_lookup_elem(&goids, &key);;
    if (!stack_id)
        return 0; // should not happen
    (*stack_id)++;
    __u32 cpu = bpf_get_smp_processor_id();
    return (*stack_id) | ((__u64)cpu << 32);
}

SEC("uprobe/on_entry_golang")
int on_entry_golang(struct pt_regs *ctx) {
    struct event this_event;
    __builtin_memset(&this_event, 0, sizeof(this_event));

    // manipulate bpf inst
    //void *a, *b;
    //bpf_probe_read_user(&a, sizeof(a), (void*)ctx->rax+8);
    //bpf_probe_read_user(&b, sizeof(b), (void*)a);
    //__builtin_memcpy(&this_event.args, &b, sizeof(b));
    // manipulation ends

    __u64 this_bp = ctx->rbp;
    this_event.hook_point = 0;
    this_event.ip = ctx->rip;
    bpf_probe_read_user(&this_event.caller_ip, sizeof(this_event.caller_ip), (void*)ctx->rbp+8);
    this_event.time_ns = bpf_ktime_get_ns();

    __u64 caller_bp;
    bpf_probe_read_user(&caller_bp, sizeof(caller_bp), (void*)ctx->rbp);
    struct event *caller_event = bpf_map_lookup_elem(&bp_to_event, &caller_bp);
    if (caller_event) {
        this_event.stack_id = caller_event->stack_id;
        this_event.stack_depth = caller_event->stack_depth + 1;
        goto submit_event;
    }

    struct stackwalk walk;
    __builtin_memset(&walk, 0, sizeof(walk));
    backtrace(ctx->rbp, &walk);
    this_event.stack_depth = walk.depth;
    this_event.stack_id = walk.stack_id;
    if (walk.depth == 0xffffffffffffffff) {
        this_event.errno = 2;
        goto submit_event;
    }

    if (this_event.stack_id == 0) {
        this_event.stack_id = new_stack_id();
        bpf_probe_write_user((void*)walk.root_bp, &this_event.stack_id, sizeof(this_event.stack_id));
        goto submit_event;
    }

submit_event:
    bpf_map_update_elem(&bp_to_event, &this_bp, &this_event, BPF_ANY);
    bpf_map_push_elem(&event_queue, &this_event, BPF_EXIST);
    return 0;
}

SEC("uprobe/on_exit_golang")
int on_exit_golang(struct pt_regs *ctx) {
    struct event this_event;
    __builtin_memset(&this_event, 0, sizeof(this_event));
    this_event.hook_point = 1;
    this_event.ip = ctx->rip;
    this_event.time_ns = bpf_ktime_get_ns();

    __u64 this_bp = ctx->rsp - 8;
    struct event *entry_event = bpf_map_lookup_elem(&bp_to_event, &this_bp);
    if (entry_event) {
        this_event.stack_id = entry_event->stack_id;
        this_event.stack_depth = entry_event->stack_depth;
        goto submit_event;
    }

    struct stackwalk walk;
    __builtin_memset(&walk, 0, sizeof(walk));
    backtrace(this_bp, &walk);
    this_event.stack_depth = walk.depth;
    this_event.stack_id = walk.stack_id;
    if (walk.depth == 0xffffffffffffffff) {
        this_event.errno = 2;
        goto submit_event;
    }

    if (this_event.stack_id == 0) {
        return 0; // dangling exit, do nothing
    }

submit_event:
    bpf_map_delete_elem(&bp_to_event, &this_bp);
    bpf_map_push_elem(&event_queue, &this_event, BPF_EXIST);
    return 0;
}

SEC("uprobe/on_entry")
int on_entry(struct pt_regs *ctx) {
    struct event this_event;
    __builtin_memset(&this_event, 0, sizeof(this_event));
    this_event.hook_point = 0;
    this_event.ip = ctx->rip;
    bpf_probe_read_user(&this_event.caller_ip, sizeof(this_event.caller_ip), (void*)ctx->rsp);
    this_event.time_ns = bpf_ktime_get_ns();

    __u64 this_bp = ctx->rsp - 8;
    bpf_printk("enter %llx, this bp %llx\n", ctx->rsp, this_bp);
    __u64 caller_bp = ctx->rbp;
    struct event *caller_event = bpf_map_lookup_elem(&bp_to_event, &caller_bp);
    if (caller_event) {
        this_event.stack_id = caller_event->stack_id;
        this_event.stack_depth = caller_event->stack_depth + 1;
        goto submit_event;
    }

    this_event.stack_id = new_stack_id();

submit_event:
    bpf_map_update_elem(&bp_to_event, &this_bp, &this_event, BPF_ANY);
    bpf_printk("push event, depth %d, ip %lld, caller bp %lld\n", this_event.stack_depth, this_event.ip, caller_bp);
    bpf_map_push_elem(&event_queue, &this_event, BPF_EXIST);
    return 0;
}

SEC("uretprobe/on_exit")
int on_exit(struct pt_regs *ctx) {
    struct event this_event;
    __builtin_memset(&this_event, 0, sizeof(this_event));
    this_event.hook_point = 1;
    this_event.time_ns = bpf_ktime_get_ns();

    __u64 this_bp = ctx->rsp - 16;
    bpf_printk("return %llx, this bp %llx\n", ctx->rsp, this_bp);
    struct event *entry_event = bpf_map_lookup_elem(&bp_to_event, &this_bp);
    if (entry_event) {
        this_event.stack_id = entry_event->stack_id;
        this_event.stack_depth = entry_event->stack_depth;
        bpf_map_delete_elem(&bp_to_event, &this_bp);
        bpf_printk("push event\n");
        bpf_map_push_elem(&event_queue, &this_event, BPF_EXIST);
    }

    return 0;
}
