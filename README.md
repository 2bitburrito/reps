# REPS
A command line tool that will list all of the repos in your organisation and allow you to clone directly

### Installation
Ensure you have installed fzf and the gh cli

#### MacOS:

```bash
brew install fzf
brew install gh
```

Then ensure you have signed in to the github cli with:
```bash
gh auth login
```

> [!NOTE]
> If you are already logged in on gh to another org you may need to first run gh switch. 
> You can see all available accounts with: `gh auth status`

### Usage
Simply run: `reps <organisation-name>`

You should see a 

