class DataNode:
    def __init__(self, name):
        self.name = name
        self.next = None

class SinglyLinkedList:
    def __init__(self):
        self.count = 0
        self.head = None

    def traverse(self):
        start = self.head
        while start != None:
            print("->", start.name, end=" ")
            start = start.next
        if self.head == None:
            print("This is an empty list.")
        print()

    def insert(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
        else:
            pNew.next = self.head
            self.head = pNew
                    

mylist = SinglyLinkedList()
mylist.insert("Tony")
mylist.traverse()
mylist.insert("Mika")
mylist.traverse()
mylist.traverse()