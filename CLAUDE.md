# Writing for One FTP Server

These rules cover the prose of this repository: the Markdown files, the release notes, and anything written to GitHub such as an issue comment, a pull request description or a release body.

## No em dashes

Never use an em dash (`—`), and never use an en dash (`–`) in its place.
Where one would have gone, use the punctuation the sentence actually calls for, or join the clauses with a conjunction.

    no    The server writes nothing — no configuration file, no key store.
    yes   The server writes nothing: no configuration file, no key store.

    no    In a `.bat` file this makes no difference — both land in the cmd window — but redirection still works.
    yes   In a `.bat` file this makes no difference, since both land in the cmd window, but redirection still works.

A colon introduces what follows, a comma or a semicolon separates clauses, and parentheses hold an aside.
One of those fits every place an em dash would have.

## One sentence, one line

In Markdown, write each sentence on a line of its own and let it run as long as it needs to.
Do not wrap a paragraph to a column width: a sentence is never broken across two lines, and two sentences never share one.
The rendered page is unaffected, and a diff then shows the sentence that changed rather than every line the rewrap moved.

    no    The certificate is generated at startup, in memory: it lasts a
          year, covers `localhost` and the addresses of the machine, and is
          replaced by a new one on every start. It is signed by nobody.

    yes   The certificate is generated at startup, in memory: it lasts a year, covers `localhost` and the addresses of the machine, and is replaced by a new one on every start.
          It is signed by nobody.

Inside a list item, continuation sentences line up with the text of the item:

    - `--log` is the file the activity goes to.
      `--log=off` keeps no file at all.

Code blocks, tables and headings are left exactly as they are; the rule is about prose.
