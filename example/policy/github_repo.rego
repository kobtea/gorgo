package github.repo

warn[msg] {
  [y, m, _, _, _, _] := time.diff(time.now_ns(), time.parse_rfc3339_ns(input.pushed_at))
  y * 12 + m > 6
  msg := "GitHub repository should be pushed at least once every 6 month"
}
