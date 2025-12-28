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

let unused = 99

let lolll = 99
let dumb = 220202002
let asdasdsaasd = 2310238
--these all are unused and should be removed by the optimizer
let foldedConst = 1/2
-- ^^ this should be folded into 0.5
print(foldedConst)
let foldedConst2 = 20
foldedConst2 = foldedConst2 * 2 -- should be folded into let foldedConst2 = 40
print(foldedConst2)
