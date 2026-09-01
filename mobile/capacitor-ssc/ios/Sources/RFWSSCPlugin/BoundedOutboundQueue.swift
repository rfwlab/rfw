import Foundation

struct BoundedOutboundQueue {
    private let capacity: Int
    private var pending: [String] = []
    private(set) var sending = false

    init(capacity: Int) {
        self.capacity = capacity
    }

    mutating func enqueue(_ value: String) -> Bool {
        let occupied = pending.count + (sending ? 1 : 0)
        guard occupied < capacity else { return false }
        pending.append(value)
        return true
    }

    mutating func next() -> String? {
        guard !sending, !pending.isEmpty else { return nil }
        sending = true
        return pending.removeFirst()
    }

    mutating func complete() {
        sending = false
    }

    mutating func removeAll() {
        pending.removeAll(keepingCapacity: false)
        sending = false
    }
}
