example = '''
class ArrayQueue:
    def __init__(self):
        self.data = []

    def size(self):
        return len(self.data)

    def is_empty(self):
        return self.data == []

    def enqueue(self, input_data):
        self.data.append(input_data)

    def dequeue(self):
        if self.data == []:
            print("Underflow: Cannot dequeue data from an empty queue")
            return None
        else:
            return self.data.pop(0)  # FIFO - remove from front

    def front(self):
        if self.data == []:
            return None
        return self.data[0]

    def back(self):
        if self.data == []:
            return None
        return self.data[-1]

    def printQueue(self):
        print(self.data)


def reverse_queue(q):
    if q.is_empty():
        return
    
    stack = []
    # Dequeue all elements and push to stack
    while not q.is_empty():
        stack.append(q.dequeue())
    
    # Pop from stack and enqueue back
    while stack:
        q.enqueue(stack.pop())


# --------------------------
# 🔹 ทดสอบโปรแกรมทั้งหมด
# --------------------------
myQueue = ArrayQueue()
myQueue.enqueue(10)
myQueue.enqueue(20)
myQueue.enqueue(30)
myQueue.printQueue()

front_value = myQueue.front()
print("Front value:", front_value)

back_value = myQueue.back()
print("Back value:", back_value)

dequeued = myQueue.dequeue()
print("Dequeued:", dequeued)
myQueue.printQueue()

print("Queue size:", myQueue.size())
print("Is empty:", myQueue.is_empty())

# Test reverse queue
q2 = ArrayQueue()
q2.enqueue(1)
q2.enqueue(2)
q2.enqueue(3)
q2.enqueue(4)
print("Before reverse:")
q2.printQueue()
reverse_queue(q2)
print("After reverse:")
q2.printQueue()
'''

