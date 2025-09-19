package com.example;

/**
 * Simple Java application for testing Gradle integration.
 */
public class App {
    
    public static void main(String[] args) {
        App app = new App();
        System.out.println(app.getGreeting());
        
        if (args.length > 0) {
            System.out.println("Arguments received: " + String.join(", ", args));
        }
    }
    
    /**
     * Returns a greeting message.
     * @return greeting string
     */
    public String getGreeting() {
        return "Hello, Gradle from mvx!";
    }
    
    /**
     * Adds two numbers.
     * @param a first number
     * @param b second number
     * @return sum of a and b
     */
    public int add(int a, int b) {
        return a + b;
    }
    
    /**
     * Checks if a string is empty or null.
     * @param str string to check
     * @return true if string is null or empty
     */
    public boolean isEmpty(String str) {
        return str == null || str.trim().isEmpty();
    }
}
