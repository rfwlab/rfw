import XCTest
@testable import RFWSSCPlugin

final class BoundedOutboundQueueTests: XCTestCase {
    func testCapacityIncludesTheInflightFrameAndPreservesOrder() {
        var queue = BoundedOutboundQueue(capacity: 2)

        XCTAssertTrue(queue.enqueue("one"))
        XCTAssertEqual(queue.next(), "one")
        XCTAssertTrue(queue.enqueue("two"))
        XCTAssertFalse(queue.enqueue("three"))

        queue.complete()
        XCTAssertEqual(queue.next(), "two")
        queue.complete()
        XCTAssertNil(queue.next())
    }

    func testRemoveAllResetsInflightAndPendingState() {
        var queue = BoundedOutboundQueue(capacity: 1)
        XCTAssertTrue(queue.enqueue("one"))
        XCTAssertEqual(queue.next(), "one")

        queue.removeAll()

        XCTAssertFalse(queue.sending)
        XCTAssertTrue(queue.enqueue("two"))
        XCTAssertEqual(queue.next(), "two")
    }
}
