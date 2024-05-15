package main

import "testing"

func TestBadWordReplacement(t *testing.T){
  got := badWordReplacement("I had something interesting for breakfast fornax kerfuffle")
  want := "I had something interesting for breakfast **** ****"
  if got != want {
    t.Errorf("got %q, wanted %q", got, want)
  }
}
