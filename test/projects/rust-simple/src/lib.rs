// Re-export main module items for testing
pub use crate::main::{Calculator, Person, filter_adults, people_from_json, people_to_json};

// Include the main module
#[path = "main.rs"]
mod main;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_person_creation() {
        let person = Person::new("Alice".to_string(), 25);
        assert_eq!(person.name, "Alice");
        assert_eq!(person.age, 25);
        assert_eq!(person.email, None);
    }

    #[test]
    fn test_person_with_email() {
        let person = Person::new("Bob".to_string(), 30)
            .with_email("bob@example.com".to_string());
        assert_eq!(person.email, Some("bob@example.com".to_string()));
    }

    #[test]
    fn test_person_is_adult() {
        let adult = Person::new("Alice".to_string(), 25);
        let minor = Person::new("Bob".to_string(), 17);
        let exactly_18 = Person::new("Charlie".to_string(), 18);

        assert!(adult.is_adult());
        assert!(!minor.is_adult());
        assert!(exactly_18.is_adult());
    }

    #[test]
    fn test_person_greet() {
        let person = Person::new("Alice".to_string(), 25);
        let greeting = person.greet();
        assert_eq!(greeting, "Hello, my name is Alice and I'm 25 years old!");
    }

    #[test]
    fn test_calculator_add() {
        assert_eq!(Calculator::add(5, 3), 8);
        assert_eq!(Calculator::add(-5, 3), -2);
        assert_eq!(Calculator::add(0, 0), 0);
    }

    #[test]
    fn test_calculator_subtract() {
        assert_eq!(Calculator::subtract(10, 3), 7);
        assert_eq!(Calculator::subtract(3, 10), -7);
        assert_eq!(Calculator::subtract(5, 5), 0);
    }

    #[test]
    fn test_calculator_multiply() {
        assert_eq!(Calculator::multiply(4, 3), 12);
        assert_eq!(Calculator::multiply(-4, 3), -12);
        assert_eq!(Calculator::multiply(0, 100), 0);
    }

    #[test]
    fn test_calculator_divide() {
        assert_eq!(Calculator::divide(10, 2), Some(5));
        assert_eq!(Calculator::divide(7, 3), Some(2)); // Integer division
        assert_eq!(Calculator::divide(10, 0), None);
        assert_eq!(Calculator::divide(0, 5), Some(0));
    }

    #[test]
    fn test_filter_adults() {
        let people = vec![
            Person::new("Alice".to_string(), 25),
            Person::new("Bob".to_string(), 17),
            Person::new("Charlie".to_string(), 18),
            Person::new("David".to_string(), 16),
        ];

        let adults = filter_adults(people);
        assert_eq!(adults.len(), 2);
        assert_eq!(adults[0].name, "Alice");
        assert_eq!(adults[1].name, "Charlie");
    }

    #[test]
    fn test_json_serialization() {
        let people = vec![
            Person::new("Alice".to_string(), 25).with_email("alice@example.com".to_string()),
            Person::new("Bob".to_string(), 30),
        ];

        let json = people_to_json(&people).expect("Serialization should succeed");
        assert!(json.contains("Alice"));
        assert!(json.contains("alice@example.com"));
        assert!(json.contains("Bob"));

        let deserialized: Vec<Person> = people_from_json(&json).expect("Deserialization should succeed");
        assert_eq!(deserialized, people);
    }

    #[test]
    fn test_empty_filter_adults() {
        let people = vec![
            Person::new("Child1".to_string(), 10),
            Person::new("Child2".to_string(), 15),
        ];

        let adults = filter_adults(people);
        assert!(adults.is_empty());
    }

    #[test]
    fn test_person_clone() {
        let person1 = Person::new("Alice".to_string(), 25);
        let person2 = person1.clone();
        
        assert_eq!(person1, person2);
        assert_eq!(person1.name, person2.name);
        assert_eq!(person1.age, person2.age);
    }
}
