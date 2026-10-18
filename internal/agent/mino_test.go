package agent

import (
	"fmt"
	"testing"
)

func TestChat(t *testing.T) {
	mino, err := newMimo()
	if err != nil {
		t.Fatal(err)
	}
	/*
		answer, err := mino.Chat("你是谁")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("\nanswer:", answer)

		answer, err = mino.Chat("你近视吗")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("\nanswer:", answer)
	*/
	answer, err := mino.Chat("世界首富是谁")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\nanswer:", answer)
}
