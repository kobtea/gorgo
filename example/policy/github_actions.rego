package github.actions

warn[msg] {
  count({x | input.jobs[_].steps[x].name == "Install dependencies"}) == 0
  msg := "GitHub actions should be defined `Install dependencies` step"
}
