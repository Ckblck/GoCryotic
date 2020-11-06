workflow "Unit Tests" {
  resolves = ["cedrickring/golang-action@1.6.0"]
  on = "push"
}

action "cedrickring/golang-action@1.6.0" {
  uses = "cedrickring/golang-action@1.6.0"
}
