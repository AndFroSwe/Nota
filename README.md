# NoTa - Note Tasks - Simple and effective task tracking 🗒️

Todos often come up while notes are being taken, for example during meetings. Instead of having a separate note taking
app which would break the flow, just embed tasks directly in the notes. NoTa can then scan all files with notes in them,
extract todos and present them in a digestible format.

## About

NoTa tracks todos by adding special tags on top of regular markdown tick box syntax `- [ ]`. Note files can be any plain
text format as long as todos to be tracked are marked correctly. Available tags are:

* `[[resp1, resp2, ...]]` Responsibles for completing the task
* `{tag1, tag2, ...}` Tags for the task
* `<yymmdd>` Deadline date for completing the task

An example of a file with notes:
```markdown
# This is a file with notes

* A bullet point
- [ ] A task to be completed <260102> by someone [[PERSON1]]
- [x] <260101> A done task
* Text between tasks is ok
- [-] A canceled task {canceled}

## Regular markdown

More text and todos...
```

## Usage

NoTa is a command line tool written in Go. Download a release or run/install it via Go's install system:

```bash
go install github.com/andfroswe/nota/cmd/nota@latest
```

Full CLI options can be checked by running `nota --help`.
