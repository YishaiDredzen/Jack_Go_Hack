// C_PUSH constant 7
@7
D=A
@SP
A=M
M=D
@SP
M=M+1
// C_PUSH constant 9
@9
D=A
@SP
A=M
M=D
@SP
M=M+1
// gt
@SP
AM=M-1
D=M
A=A-1
D=M-D
@TRUE_0
D;JGT
@SP
A=M-1
M=0
@END_0
0;JMP
(TRUE_0)
@SP
A=M-1
M=-1
(END_0)
