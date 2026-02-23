example = '''
class DataNode:
    ### Create Data ###
    def __init__(self, name):
        self.name = name
        self.next = None
        self.prev = None

class DoublyLinkedList:
    ### Create List ###
    def __init__(self):
        self.count = 0
        self.head = None
        self.tail = None

    ### Print all name ###
    def traverse(self):
        start = self.head
        while start != None:
            print("->", start.name, end=" ")
            start = start.next
        if self.head == None:
            print("This is an empty list.")
        print()
    
    ### Print all name in reverse ###
    def traverseReverse(self):
        start = self.tail
        while start != None:
            print("->", start.name, end=" ")
            start = start.prev
        if self.tail == None:
            print("This is an empty list.")
        print()
    
    ### Insert data at the front ###
    def insertFront(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
            self.tail = pNew
        else:
            pNew.next = self.head
            self.head.prev = pNew
            self.head = pNew

    ### Insert Data at the end ###
    def insertLast(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
            self.tail = pNew
        else:
            self.tail.next = pNew
            pNew.prev = self.tail
            self.tail = pNew
    
    ### Insert Data Between Data ###
    def insertBefore(self, Node, name):
        pNew = DataNode(name)
        if self.head == None:
            print("Cannot insert, list is empty.")
            return
            
        if self.head.name == Node:
            pNew.next = self.head
            self.head.prev = pNew
            self.head = pNew
        else:
            start = self.head
            while start.next != None:
                if start.next.name == Node:
                    pNew.next = start.next
                    pNew.prev = start
                    start.next.prev = pNew
                    start.next = pNew
                    return
                start = start.next
            print("Cannot insert, <" + Node + "> does not exist.")

    def delete(self, name):
        if self.head == None:
            print("Cannot delete, list is empty.")
            return
            
        if self.head.name == name:
            if self.head == self.tail:  # Only one node
                self.head = None
                self.tail = None
            else:
                self.head = self.head.next
                self.head.prev = None
        else:
            start = self.head
            while start.next != None:
                if start.next.name == name:
                    if start.next == self.tail:  # Deleting last node
                        self.tail = start
                        start.next = None
                    else:
                        start.next = start.next.next
                        start.next.prev = start
                    return
                start = start.next
            print("Cannot delete, <" + name + "> does not exist.")
                    

mylist = DoublyLinkedList()
mylist.insertFront("Tony")
mylist.insertFront("John")
mylist.traverse()
mylist.traverseReverse()
mylist.insertFront("Mika")
mylist.insertLast("Saori")
mylist.insertBefore("John", "Ako")
mylist.traverse()
mylist.traverseReverse()
mylist.delete("John")
mylist.delete("Tony")
mylist.insertBefore("Saori", "Yaoyao")
mylist.traverse()
mylist.traverseReverse()
'''

