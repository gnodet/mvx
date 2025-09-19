package com.example;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for the App class.
 */
public class AppTest {
    
    private App app;
    
    @BeforeEach
    void setUp() {
        app = new App();
    }
    
    @Test
    @DisplayName("Should return correct greeting message")
    void testGetGreeting() {
        String greeting = app.getGreeting();
        assertEquals("Hello, Gradle from mvx!", greeting);
        assertNotNull(greeting);
        assertFalse(greeting.isEmpty());
    }
    
    @Test
    @DisplayName("Should add two positive numbers correctly")
    void testAddPositiveNumbers() {
        int result = app.add(5, 3);
        assertEquals(8, result);
    }
    
    @Test
    @DisplayName("Should add negative numbers correctly")
    void testAddNegativeNumbers() {
        int result = app.add(-5, -3);
        assertEquals(-8, result);
    }
    
    @Test
    @DisplayName("Should add zero correctly")
    void testAddWithZero() {
        assertEquals(5, app.add(5, 0));
        assertEquals(0, app.add(0, 0));
        assertEquals(-5, app.add(-5, 0));
    }
    
    @Test
    @DisplayName("Should detect empty strings correctly")
    void testIsEmpty() {
        assertTrue(app.isEmpty(null));
        assertTrue(app.isEmpty(""));
        assertTrue(app.isEmpty("   "));
        assertTrue(app.isEmpty("\t\n"));
        
        assertFalse(app.isEmpty("hello"));
        assertFalse(app.isEmpty(" hello "));
        assertFalse(app.isEmpty("0"));
    }
}
