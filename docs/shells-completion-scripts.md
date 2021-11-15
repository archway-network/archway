# Shells Completion Scripts

Completion scripts for popular UNIX shell interpreters such as `Bash` and `Zsh`
can be generated through the `completion` command, which is available for `archwayd`.

If you want to generate `Bash` completion scripts run the following command:

```bash
cd ~
archwayd completion > archwayd_completion
```

If you want to generate `Zsh` completion scripts run the following command:

```bash
archwayd completion --zsh > archwayd_completion
```

**Warning:** On some linux OS, you may face the following error:

```sh
$'\r': command not found
```
We need to remove the carriage return character as it causes that issue.

```bash
sed 's/\r$//' archwayd_completion > archwayd_completion
```

>**Tip:**
On most UNIX systems, such scripts may be loaded in `.bashrc` or
`.bash_profile` to enable Bash autocompletion:

```bash
echo '. archwayd_completion' >> ~/.bashrc
```

Refer to the user's manual of your interpreter provided by your
operating system for information on how to enable shell autocompletion.

For the updated version of this document, please refer to the [original version](https://github.com/cosmos/gaia/blob/main/docs/resources/gaiad.md#shells-completion-scripts).