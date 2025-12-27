let start = tick()

let i = 0
while i < 10000000 do
    i = i + 1
end

let end_time = tick()
let elapsed = end_time - start
-- print out results ((( test if + then <-- this was a test if the parser is dumb)))
print("Sum: " + i)
print("Elapsed time: " + elapsed + " seconds") -- another test at parser
print("Iterations per second: " + 10000000 / elapsed) -- fixed
-- anonymous function
let add = func(z, y) return z + y end

print("Anonymous function test: " + add(5, 3))

-- inline anonymous function
print("Inline function: " + func(a) print(a) end("hi"))
