# M1 C3: headings present in the initial SSR body markup (title/script/style/template excluded). Diagnostic only.
import sys
from html.parser import HTMLParser
SKIP = {'script', 'style', 'template', 'title', 'head'}
class P(HTMLParser):
    def __init__(self):
        super().__init__(); self.skip = 0; self.h = None; self.buf = ''; self.out = []
    def handle_starttag(self, tag, attrs):
        if tag in SKIP: self.skip += 1
        elif not self.skip and tag in ('h1', 'h2', 'h3'): self.h, self.buf = tag, ''
    def handle_endtag(self, tag):
        if tag in SKIP: self.skip = max(0, self.skip - 1)
        elif self.h == tag: self.out.append((tag, ' '.join(self.buf.split()))); self.h = None
    def handle_data(self, d):
        if self.h and not self.skip: self.buf += d
p = P(); p.feed(open(sys.argv[1], encoding='utf-8').read())
want = [tuple(w.split('=', 1)) for w in sys.argv[2:]]  # e.g. h1=Каталог курсів
missing = [w for w in want if w not in p.out]
print('headings:', p.out[:6], '| want:', want, '| ok' if not missing else f'| MISSING {missing}')
sys.exit(1 if missing else 0)
