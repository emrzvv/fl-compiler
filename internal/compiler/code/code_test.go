package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []int
		expected []byte
	}{
		{
			OpConstant,
			[]int{65534},
			[]byte{byte(OpConstant), 255, 254},
		},
		{
			OpAdd,
			[]int{2},
			[]byte{byte(OpAdd), 0, 2},
		},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		t.Logf("%d", instruction)
		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d", len(tt.expected), len(instruction))
		}

		for i, b := range tt.expected {
			if instruction[i] != tt.expected[i] {
				t.Errorf("wrong byte at pos %d. expected %d, got %d", i, b, instruction[i])
			}
		}
	}

}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd, 2),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd 2
0003 OpConstant 2
0006 OpConstant 65535
`

	concatted := Instructions{}
	for _, instr := range instructions {
		concatted = append(concatted, instr...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nexpected %q\ngot %q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        OpCode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}
	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
		}
		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
