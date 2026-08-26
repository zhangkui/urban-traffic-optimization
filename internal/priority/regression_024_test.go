package priority

import ("testing"; "time")
func TestDecideWithContextRejectsCancelledRequest(t *testing.T) { s:=NewService(); if err:=s.Submit(Request{ID:"priority-24",IntersectionID:"i24",Direction:"north",Vehicle:"ambulance",RequestedAt:time.Now().Add(-time.Hour),ExpiresAt:time.Now().Add(-time.Minute)});err!=nil{t.Fatal(err)}; got:=s.Decide("priority-24",time.Now());if got.Granted{t.Fatalf("过期请求仍获授权=%+v",got)} }
