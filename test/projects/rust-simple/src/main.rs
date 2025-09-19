use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// A simple struct to demonstrate Rust features
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Person {
    pub name: String,
    pub age: u32,
    pub email: Option<String>,
}

impl Person {
    /// Creates a new Person
    pub fn new(name: String, age: u32) -> Self {
        Person {
            name,
            age,
            email: None,
        }
    }

    /// Sets the email for the person
    pub fn with_email(mut self, email: String) -> Self {
        self.email = Some(email);
        self
    }

    /// Checks if the person is an adult (18 or older)
    pub fn is_adult(&self) -> bool {
        self.age >= 18
    }

    /// Returns a greeting message
    pub fn greet(&self) -> String {
        format!("Hello, my name is {} and I'm {} years old!", self.name, self.age)
    }
}

/// A simple calculator struct
#[derive(Debug)]
pub struct Calculator;

impl Calculator {
    /// Adds two numbers
    pub fn add(a: i32, b: i32) -> i32 {
        a + b
    }

    /// Subtracts two numbers
    pub fn subtract(a: i32, b: i32) -> i32 {
        a - b
    }

    /// Multiplies two numbers
    pub fn multiply(a: i32, b: i32) -> i32 {
        a * b
    }

    /// Divides two numbers, returns None if dividing by zero
    pub fn divide(a: i32, b: i32) -> Option<i32> {
        if b == 0 {
            None
        } else {
            Some(a / b)
        }
    }
}

/// Processes a list of people and returns adults only
pub fn filter_adults(people: Vec<Person>) -> Vec<Person> {
    people.into_iter().filter(|p| p.is_adult()).collect()
}

/// Converts people to JSON string
pub fn people_to_json(people: &[Person]) -> Result<String, serde_json::Error> {
    serde_json::to_string_pretty(people)
}

/// Parses people from JSON string
pub fn people_from_json(json: &str) -> Result<Vec<Person>, serde_json::Error> {
    serde_json::from_str(json)
}

fn main() {
    println!("🦀 Hello from Rust via mvx!");

    // Create some sample people
    let people = vec![
        Person::new("Alice".to_string(), 25).with_email("alice@example.com".to_string()),
        Person::new("Bob".to_string(), 17),
        Person::new("Charlie".to_string(), 30).with_email("charlie@example.com".to_string()),
    ];

    println!("\nAll people:");
    for person in &people {
        println!("  {}", person.greet());
    }

    // Filter adults
    let adults = filter_adults(people.clone());
    println!("\nAdults only:");
    for adult in &adults {
        println!("  {}", adult.greet());
    }

    // Demonstrate calculator
    println!("\nCalculator demo:");
    println!("  10 + 5 = {}", Calculator::add(10, 5));
    println!("  10 - 5 = {}", Calculator::subtract(10, 5));
    println!("  10 * 5 = {}", Calculator::multiply(10, 5));
    
    match Calculator::divide(10, 5) {
        Some(result) => println!("  10 / 5 = {}", result),
        None => println!("  10 / 5 = Error: Division by zero"),
    }

    match Calculator::divide(10, 0) {
        Some(result) => println!("  10 / 0 = {}", result),
        None => println!("  10 / 0 = Error: Division by zero"),
    }

    // JSON serialization demo
    if let Ok(json) = people_to_json(&adults) {
        println!("\nAdults as JSON:");
        println!("{}", json);
    }
}
