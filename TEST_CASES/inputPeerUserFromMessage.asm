SUB SP, SP, #0x40
STP X22, X21, [SP,#16]
STP X20, X19, [SP,#32]
STP X29, X30, [SP,#48]
ADD X29, SP, #0x30
ADRP X8, .+0x4c9000
LDR X8, [X8,#1272]
LDR X8, [X8]
STR X8, [SP,#8]
LSR X8, X2, #61
MOV X20, X0
ADRP X9, .+0x38c000
ADD X9, X9, #0xd34
ADR X10, .+0x10
LDRB W11, [X9,X8]
ADD X10, X10, X11, LSL #2
BR X10
LDP X21, X19, [X2,#16]
TBZ W1, #0, .+0x6c (inputPeerChannel)
MOV W8, #0xbbfc
MOVK W8, #0x27bc, LSL #16
B .+0x50
LDR X19, [X2,#16]
LDR W22, [X2,#24]
LDR X21, [X2,#32]
TBZ W1, #0, .+0x90 (inputPeerChannelFromMessage)
MOV W8, #0x840
MOVK W8, #0xbd2a, LSL #16
B .+0x74
LDR X19, [X2,#16]
TBZ W1, #0, .+0x4c (inputPeerChat)
MOV W8, #0x5cb9
MOVK W8, #0x35a9, LSL #16
STR W8, [SP]
MOV X0, SP
MOV W1, #0x4
B .+0x30
LDP X21, X19, [X2,#16]
TBZ W1, #0, .+0x1c (inputPeerUser)
MOV W8, #0xa54c
MOVK W8, #0xdde8, LSL #16
STR W8, [SP]
MOV X0, SP
MOV W1, #0x4
BL .+0x3021d8
STR X21, [SP]
MOV X0, SP
MOV W1, #0x8
BL .+0x3021c8
STR X19, [SP]
B .+0x50
LDR X19, [X2,#16]
LDR W22, [X2,#24]
LDR X21, [X2,#32]
TBZ W1, #0, .+0x1c (inputPeerUserFromMessage)
MOV W8, #0xa1c
MOVK W8, #0xa87b, LSL #16
STR W8, [SP]
MOV X0, SP
MOV W1, #0x4
BL .+0x302198
MOV X0, X20
MOV W1, #0x1
MOV X2, X19
BL .+0xffffffffffffff00
STR W22, [SP]
MOV X0, SP
MOV W1, #0x4
BL .+0x302178
STR X21, [SP]
MOV X0, SP
MOV W1, #0x8
BL .+0x302168
LDR X8, [SP,#8]
ADRP X9, .+0x4c9000
LDR X9, [X9,#1272]
LDR X9, [X9]
CMP X9, X8
B NE, .+0x50
LDP X29, X30, [SP,#48]
LDP X20, X19, [SP,#32]
LDP X22, X21, [SP,#16]
ADD SP, SP, #0x40
RET X30
MOV X8, #0xa000000000000000
CMP X2, X8
B NE, .+0x14
TBZ W1, #0, .+0xffffffffffffffc8 (inputPeerEmpty)
MOV W8, #0x18ea
MOVK W8, #0x7f3b, LSL #16
B .+0x10
TBZ W1, #0, .+0xffffffffffffffb8
MOV W8, #0x7ec9
MOVK W8, #0x7da0, LSL #16 (inputPeerSelf)
STR W8, [SP]
MOV X0, SP
MOV W1, #0x4
B .+0xffffffffffffff9c
BL .+0x3471b8