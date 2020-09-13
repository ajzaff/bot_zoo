package zoo

// position values for Color > Piece > Square.
// Squares are in order from A1, A2, ... to H8.
var positionValue = [][][]Value{{{}, { // GRabbit
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 0, -5, -5, 0, 1, 1,
	1, 2, -3, -3, -3, -3, 2, 1,
	2, 1, 0, -3, -3, 0, 1, 2,
	2, 0, -3, -3, -3, -3, 0, 2,
	5, -2, -10, -5, -5, -10, -2, 5,
	10, 10, 5, 10, 10, 5, 10, 10,
	999, 999, 999, 999, 999, 999, 999, 999,
}, { // GCat
	5, 10, 10, 5, 5, 10, 10, 5,
	5, 8, 10, 5, 5, 10, 8, 10,
	0, 0, -2, 0, 0, -2, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // GDog
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // GHorse
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, 0, 0, 5, 0, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // GCamel
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 0, 0, 5, 5, 0, 0, -8,
	-8, 0, 0, 2, 2, 0, 0, -8,
	-8, 0, 0, 2, 2, 0, 5, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // GElephant
	-22, -13, -13, -13, -13, -13, -13, -22,
	-13, 0, 0, 0, 0, 0, 0, -13,
	-13, 5, -8, 10, 10, -8, 5, -13,
	-13, 5, 10, 10, 10, 10, 5, -13,
	-13, 5, 5, 10, 10, 5, 5, -13,
	-13, 5, 0, 5, 5, 0, 5, -13,
	-13, 0, -1, 0, 0, -1, 0, -13,
	-22, -13, -13, -13, -13, -13, -13, -22,
}}, {{}, { // SRabbit
	999, 999, 999, 999, 999, 999, 999, 999,
	10, 10, 5, 10, 10, 5, 10, 10,
	5, -2, -10, -5, -5, -10, -2, 5,
	2, 0, -3, -3, -3, -3, 0, 2,
	2, 1, 0, -3, -3, 0, 1, 2,
	1, 2, -3, -3, -3, -3, 2, 1,
	1, 1, 0, -5, -5, 0, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
}, { // SCat
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -2, 0, 0, -2, 0, 0,
	5, 8, 10, 5, 5, 10, 8, 10,
	5, 10, 10, 5, 5, 10, 10, 5,
}, { // SDog
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, -8, -10, -8, -8, -10, -8, 0,
	0, 0, -8, 0, 0, -8, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 10, 10, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}, { // SHorse
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 10, 2, 2, 10, 5, -8,
	-8, 5, 5, 2, 2, 5, 5, -8,
	-8, 5, 0, 0, 5, 0, 5, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // SCamel
	-13, -8, -8, -8, -8, -8, -8, -13,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-8, 5, -8, 10, 10, -8, 5, -8,
	-8, 5, 0, 2, 2, 0, 5, -8,
	-8, 0, 0, 2, 2, 0, 0, -8,
	-8, 0, 0, 5, 5, 0, 0, -8,
	-8, 0, 0, 0, 0, 0, 0, -8,
	-13, -8, -8, -8, -8, -8, -8, -13,
}, { // SElephant
	-22, -13, -13, -13, -13, -13, -13, -22,
	-13, 0, -1, 0, 0, -1, 0, -13,
	-13, 5, 0, 5, 5, 0, 5, -13,
	-13, 5, 5, 10, 10, 5, 5, -13,
	-13, 5, 10, 10, 10, 10, 5, -13,
	-13, 5, -8, 10, 10, -8, 5, -13,
	-13, 0, 0, 0, 0, 0, 0, -13,
	-22, -13, -13, -13, -13, -13, -13, -22,
}}}