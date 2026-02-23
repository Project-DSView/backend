
import urllib.request
import json
import os

url = "http://localhost:8000/api/playground/run"
api_key = os.environ.get("API_KEY", "api_key_dsview_password_65")  # Use correct default key
headers = {
    "Content-Type": "application/json",
    "dsview-api-key": api_key
}

# The actual code from stack.py that causes the issue
code = """
class ArrayStack:
    def __init__(self):
        self.data = []

    def size(self):
        return len(self.data)

    def is_empty(self):
        return self.data == []

    def push(self, input_data):
        self.data.append(input_data)

    def pop(self):
        if self.data == []:
            print("Underflow: Cannot pop data from an empty list")
            return None
        else:
            return self.data.pop()

    def stackTop(self):
        if self.data == []:
            return None
        return self.data[-1]

    def printStack(self):
        print(self.data)


def is_parentheses_matching(expression):
    myStack = ArrayStack()
    for i in expression:
        if i == "(":
            myStack.push(i)
        elif i == ")":
            if myStack.is_empty():
                print("Parentheses in " + expression + " are unmatched")
                return False
            elif myStack.stackTop() == "(":
                myStack.pop()
            else:
                myStack.push(i)

    if myStack.size() == 0:
        return True
    else:
        print("Parentheses in " + expression + " are unmatched")
        return False


def copyStack(s1, s2):
    newStack = ArrayStack()
    while s1.size() != 0:
        data = s1.pop()
        newStack.push(data)
    while s2.size() != 0:
        s2.pop()
    while newStack.size() != 0:
        data = newStack.pop()
        s1.push(data)
        s2.push(data)
    del newStack


def infixToPostfix(expression):
    text = ""
    getStack = ArrayStack()
    prec = {'+': 1, '-': 1, '*': 2, '/': 2}

    for i in expression:
        if i.isalpha():
            text += i
        else:
            while (not getStack.is_empty() and 
                   prec.get(i, 0) <= prec.get(getStack.stackTop(), 0)):
                text += getStack.pop()
            getStack.push(i)

    while not getStack.is_empty():
        text += getStack.pop()

    return text

# Runs
newStack = ArrayStack()
expr = "(((A-B)*C))"
result = is_parentheses_matching(expr)
print(result)

s1 = ArrayStack()
s1.push(10)
s1.push(20)
s1.push(30)

s2 = ArrayStack()
s2.push(15)

copyStack(s1, s2)
s1.printStack()
s2.printStack()

s1.push(50)
s1.printStack()

exp = "A+B*C-D/E"
postfix = infixToPostfix(exp)
print("Postfix of", exp, "is", postfix)
"""

data = {
    "code": code,
    "dataType": "stack"
}

print(f"Sending request to {url}")
req = urllib.request.Request(url, data=json.dumps(data).encode('utf-8'), headers=headers)

try:
    with urllib.request.urlopen(req) as response:
        result = json.loads(response.read().decode('utf-8'))
        
        # Check raw stdout
        print(f"RAW STDOUT from Response: {repr(result.get('output', 'N/A'))}")
        
        steps = result.get("steps", [])
        print(f"Total steps: {len(steps)}")
        for i, step in enumerate(steps):
             state = step.get("state", {})
             if state.get("print_output"):
                 print(f"Step {i+1} Output: {state.get('print_output')}")
            
except urllib.error.HTTPError as e:
    print(f"HTTP Error: {e.code} - {e.read().decode('utf-8')}")
except Exception as e:
    print(f"Error: {e}")
