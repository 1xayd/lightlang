let sys_args = args()
let result = input("Say your name:")
print("Hi! " + result)
print(sys_args) -- args() returns an array with system arguments that were passed to the executable.
--If the program was ran through cli of the executable there will be always atleast 2 arguments which are [run location of .ll file]
