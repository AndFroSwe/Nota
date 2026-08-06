# Plan file

- [ ] Add priority tag to tasks
- [ ] Add support for setting tags on non-todo items
- [x] Add filtering
- [ ] Add coloring of deadline based on date
- [ ] File level filtering of tags
- [ ] TUI
- [x] Display column with task file
- [ ] Be able to clone git repos in Nota to keep track of multiple todo lists

## TUI/GUI

If a more advanced interface is added, add possibility of checking tasks directly in TUI/GUI, along with filtering and sorting etc.

## Backend

Keep a backend that syncs with one or more git repos. Works for small, single user use cases. Need to differ between
pure todo repos and repos where the todos are buried between other files (e.g. a code repo). The key factors are:

* Read and/or Write - Code repo -> Read, Pure todo repo -> Read + Write
* Commit history pollution - Code repo -> Don't pollute the git history, Pure todo repo -> Can add commits with changes and have a time based squash strategy
