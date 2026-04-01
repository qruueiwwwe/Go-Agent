#!/usr/bin/env python3

"""
Simple calculator module for testing
"""

class Calculator:
    """A simple calculator class"""
    
    def __init__(self):
        self.result = 0
    
    def add(self, a, b):
        """Add two numbers"""
        self.result = a + b
        return self.result
    
    def multiply(self, a, b):
        """Multiply two numbers"""
        self.result = a * b
        return self.result
    
    def get_result(self):
        """Get the current result"""
        return self.result


def main():
    calc = Calculator()
    print(f"2 + 3 = {calc.add(2, 3)}")
    print(f"5 * 4 = {calc.multiply(5, 4)}")
    print(f"Current result: {calc.get_result()}")


if __name__ == "__main__":
    main()
