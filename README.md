# Silver
The Silver interpreted programming language

Silver uses newlines to separate statements, so semicolons are unnecessary:

```silver
struct Person {
    name: string
    age: int
}

let birthday = fn(person: Person): int {
    person.age + 1
}

let ada: Person = Person("Ada", 36)
birthday(ada)
```
