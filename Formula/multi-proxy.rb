class MultiProxy < Formula
  desc "A lightweight command-line tool to spin up multiple reverse proxies with logging and debugging support."
  homepage "https://github.com/udayangaac/multi-proxy"
  url "https://github.com/udayangaac/multi-proxy/releases/download/v1.0.2/multi-proxy"
  sha256 "d3438dbc7f3ccd1516fc7ab35f8a35bd98ffe114b4024bcfbe7380e9ae9cadc5"
  version "1.0.2"

  depends_on :macos

  def install
    bin.install "multi-proxy"
  end

  test do
    output = shell_output("#{bin}/multi-proxy --help", 1)
    assert_match "Usage", output
  end
end
