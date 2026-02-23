example = '''
class DataNode:
    ### Creat Data ###
    def __init__(self, name):
        self.name = name
        self.next = None

class SinglyLinkedList:
    ### Create List ###
    def __init__(self):
        self.count = 0
        self.head = None

    ### Print all name ###
    def traverse(self):
        start = self.head
        while start != None:
            print("->", start.name, end=" ")
            start = start.next
        if self.head == None:
            print("This is an empty list.")
        print()
    
    ### Insert data at the front ###
    def insertFront(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
        else:
            pNew.next = self.head
            self.head = pNew

    ### Insert Data at the end ###
    def insertLast(self, name):
        pNew = DataNode(name)
        start = self.head
        while start.next != None:
            start = start.next
        start.next = pNew
    
    ### Insert Data Between Data ###
    def insertBefore(self, Node, name):
        pNew = DataNode(name)
        start = self.head
        if self.head.name == Node:
            pNew.next = self.head
            self.head = pNew
            if self.head == None:
                print("Cannot insert, <" + name + "> dose not exist.")
        else:
            while start.next != None:
                if start.next.name == Node:
                    pNew.next = start.next
                    start.next = pNew
                    return
                start = start.next
            print("Cannot insert, <" + Node + "> dose not exist.")

    def delete(self, name):
        if self.head.name == name:
            self.head = self.head.next
            if self.head == None:
                print("Cannot insert, <" + name + "> dose not exist.")
        else:
            start = self.head
            while start.next != None:
                if start.next.name == name:
                    start.next = start.next.next
                    return
                start = start.next
            print("Cannot delete, <" + name + "> dose not exist.")
                    

mylist = SinglyLinkedList()
mylist.insertFront("Tony")
mylist.insertFront("John")
mylist.traverse()
mylist.insertFront("Mika")
mylist.insertLast("Saori")
mylist.insertBefore("John", "Ako")
mylist.traverse()
mylist.delete("John")
mylist.delete("Tony")
mylist.insertBefore("Saori", "Yaoyao")
mylist.traverse()
'''