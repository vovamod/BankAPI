# BankAPI

[![rcard](https://goreportcard.com/badge/github.com/vovamod/BankAPI)](https://goreportcard.com/report/github.com/vovamod/BankAPI)

A Bank like system with transactions that represents each user interaction and Account with User. Account must be linked to User.
User may create as many accounts as he wants and save values on them as he pleases. (in other project that uses this one as backend, we implemented max 3 accounts per User)

# Tasks
- [x] Create a basic Server wrapper and load routes from data.go
- [x] Implement some routing via func pass to method Init() for better codebase
- [x] Add Account and User tables with CRUD ops as well
- [x] Add middleware to verify requests and their rights
- [x] Auth system (IDK, not needed nor requested, but maybe I'll implement that)
- [ ] Fix bugs with mongo (?)

# Usage
TODO