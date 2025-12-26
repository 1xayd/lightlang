let a = 10
let b = 5.5
let message = "Hello, Lightlang!"
let is_true = true
let is_false = false

let sum = a + b
let diff = a - b
let product = a * b
let quotient = a / 2

let greeting = "Hello" + " " + "World"
let mixed_concat = "Number: " + a

print "-------------------------------"
print(sum)
print(diff)
print(product)
print(quotient)
print(0.00943579734)
print(greeting)
print(mixed_concat)
print(message)
print("-------------------------------")

if is_true == is_false then
    print("not gonna print hi")
end

if is_true ~= is_false then
    print("gonna print hi")
end
print("-------------------------------")

func greet(name)
    let message = "Hello, " + name + "!"
    print(message)
    return message
end

let result = "returned message was: " + greet("alice")
print(result)
print("result len" + len(result))
print("-------------------------------")

print("random: " + random(1, 99))
print("tick: " + tick())
print("year: " + date()["year"])
print("-------------------------------")
let table = { "phrase": "meow" }

print("phrase: " + table["phrase"])
print("keys: " + keys(table))

let array = ["notmeow", "array index 2"]
print(array)
print("-------------------------------")
print("waiting uan second!!")
wait(1)
print("waited uan secand!")
print("-------------------------------")
let number1 = 1
print("type of number 1 is " + type(number1) + "!")
print("converting it to string...")
number1 = tostring(number1)
print("the type is now: " + type(number1) + "!")
print("-------------------------------")
func gamble()
    if random(1, 10) <= 4 then
        return true
    else
        return false
    end
end

print("gambling is bad but here it is anyway... Did you win a lottery? " + gamble())
print("-------------------------------")
