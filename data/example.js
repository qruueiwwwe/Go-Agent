/**
 * Simple calculator module for testing
 */

class Calculator {
    /**
     * Initialize the calculator
     */
    constructor() {
        this.result = 0;
    }

    /**
     * Add two numbers
     * @param {number} a - First number
     * @param {number} b - Second number
     * @returns {number} The sum
     */
    add(a, b) {
        this.result = a + b;
        return this.result;
    }

    /**
     * Multiply two numbers
     * @param {number} a - First number
     * @param {number} b - Second number
     * @returns {number} The product
     */
    multiply(a, b) {
        this.result = a * b;
        return this.result;
    }

    /**
     * Get the current result
     * @returns {number} The current result
     */
    getResult() {
        return this.result;
    }
}

// Example usage
const calc = new Calculator();
console.log(`2 + 3 = ${calc.add(2, 3)}`);
console.log(`5 * 4 = ${calc.multiply(5, 4)}`);
console.log(`Current result: ${calc.getResult()}`);

module.exports = Calculator;
