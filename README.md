# assembler
The virtual-machine, although far from finished, has gotten so far that an assembler is needed to write more interesting test programs. 
It now has several flow control operations and counting the bytes to know where to jump to, as well as counting the bytes to know where to obtain
a variable from is getting really tedious and error phrone.

# startup and options
The initial version is very simple. Just type `asm <filename>` and it spews out the results on stdout. In the initial version it will just be the
source code, nicely formatted followed by the byte code it thinks it needs to generate.

# assembler features
Each operation has to be on a seperate line. A regular line of code looks like: `<label>: <opcode> [<operant>]`. The possible opcodes can be found in the 
documentation of the virtual-machine. For a label you can use a valid identifier, starting with a letter or underscore and followed by up to 63 letters, 
digits, underscores or dashes.

