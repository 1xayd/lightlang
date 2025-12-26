# lightlang
A simple experiment language written in 2 days (of editor time) in golang.
I thought it'd be nice to make something like luau typescript python and other languages in golang so i decided to make this,
the 2 days thing is a lie there were earlier versions of this project that are terrible and cant even run anything unlike this one. This is the best iteration of this project yet.


To build your own version of the project use build.bat file:
```
.\build.bat 
```


To get compiled bytecode of your files:
```
lightlang build example.ll
```
This will output .llbytecode file with the name of the original file e.g:
```
'example.ll' -> 'example.llbytecode'
```


To run your files directly:
```
lightlang run example.ll
```
