package ssh

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/etkecc/go-ansible"
	"github.com/etkecc/go-kit/crypter"

	"github.com/etkecc/inventory-ssh/internal/logger"
)

var legitExitCode = map[int]bool{
	0:   true, // normal exit
	130: true, // Ctrl+C
}

// sshCrypter is a cached crypter instance for decrypting passwords and private keys, it is initialized when the first decryption is performed, it is set to nil if the secret is not provided or if there is an error creating the crypter instance, subsequent calls to getCrypter will return the cached instance or nil without creating a new one
var sshCrypter *crypter.Crypter

// Run executes the ssh command
func Run(sshCmd string, host *ansible.Host, strict bool, environ []string) {
	env := append(os.Environ(), environ...)
	cmd := buildCMD(sshCmd, host, strict, env)
	defer cleanupTempFiles(cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env

	err := cmd.Start()
	if err != nil {
		logger.Fatal("cannot start the command:", err)
	}
	err = cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && legitExitCode[exitErr.ExitCode()] {
			return
		}
		logger.Fatal("command failed:", err)
	}
}

// buildCMD builds the ssh command based on the host information and the original os.Args, it also decrypts the private keys if they are encrypted and creates temporary files for them, the temporary file paths are added to the ssh arguments, it returns the final ssh command
func buildCMD(sshCmd string, host *ansible.Host, strict bool, env []string) *exec.Cmd {
	osArgs := os.Args[1:]
	sshArgs := make([]string, 0)
	parts := strings.Split(sshCmd, " ")
	if len(parts) > 1 {
		sshCmd = parts[0]
		sshArgs = parts[1:]
	}

	if host == nil {
		if strict {
			logger.Fatal("host not found within inventory")
		}
		sshArgs = append(sshArgs, osArgs...)
		logger.Debug("command:", sshCmd, sshArgs)
		return exec.Command(sshCmd, sshArgs...) //nolint:gosec // comand is defined in config file, intended
	}

	args := buildArgs(sshArgs, osArgs, host, env)
	logger.Debug("command:", sshCmd, args)

	if host.SSHPass != "" {
		logger.Println("ssh password is:", decrypt(host.SSHPass, env))
	}

	if host.BecomePass != "" && host.User != "root" {
		logger.Println("become password is:", decrypt(host.BecomePass, env))
	}
	return exec.Command(sshCmd, args...) //nolint:gosec // that's intended
}

// buildArgs builds the ssh command arguments based on the host information and the original os.Args, it also decrypts the private keys if they are encrypted and creates temporary files for them, the temporary file paths are added to the ssh arguments, it returns the final ssh arguments
func buildArgs(sshArgs, osArgs []string, host *ansible.Host, env []string) []string {
	if host == nil {
		return nil
	}
	if sshArgs == nil {
		sshArgs = make([]string, 0)
	}

	for _, key := range host.PrivateKeys {
		keypath := decryptFile(key, env)
		sshArgs = append(sshArgs, "-i", keypath)
	}

	if host.Port != 0 {
		sshArgs = append(sshArgs, "-p", strconv.Itoa(host.Port))
	}

	if host.User != "" {
		sshArgs = append(sshArgs, host.User+"@"+host.Host)
	}

	if len(osArgs) > 1 {
		sshArgs = append(sshArgs, osArgs[1:]...)
	}

	return sshArgs
}

// getCrypter returns a crypter instance if the secret is set, otherwise it returns nil, the crypter instance is cached for subsequent calls to avoid creating multiple instances
func getCrypter(environ []string) *crypter.Crypter {
	if sshCrypter != nil {
		return sshCrypter
	}

	var err error
	prefix := "ETKE_INV_SECRET="
	for _, env := range environ {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		key := strings.TrimPrefix(env, prefix)
		sshCrypter, err = crypter.New(key)
		if err != nil {
			logger.Debug("cannot create crypter with the provided key:", err)
			return nil
		}
	}

	return nil
}

// decrypt decrypts the password if it's encrypted, otherwise it returns the original password
func decrypt(password string, environ []string) string {
	c := getCrypter(environ)
	if c == nil {
		return password
	}
	decrypted, err := c.Decrypt(password)
	if err != nil {
		logger.Fatal("cannot decrypt the password:", err)
	}
	return decrypted
}

// decryptFile decrypts the file content if it's encrypted, otherwise it returns the original file path, it creates a temporary file for the decrypted content and returns the temporary file path, the temporary file will be removed after the ssh command is executed
func decryptFile(filepath string, environ []string) string {
	c := getCrypter(environ)
	if c == nil {
		return filepath
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		logger.Fatal("cannot find the file:", filepath)
		return filepath
	}
	content, err := os.ReadFile(filepath)
	if err != nil {
		logger.Fatal("cannot read the file:", err)
		return filepath
	}
	if !c.IsEncrypted(string(content)) {
		return filepath
	}

	decrypted, err := c.Decrypt(string(content))
	if err != nil {
		logger.Fatal("cannot decrypt the file content:", err)
		return filepath
	}
	parent := path.Dir(filepath)
	tmp, err := os.CreateTemp(parent, ".inv-ssh-*")
	if err != nil {
		logger.Fatal("cannot create a temporary file:", err)
		return filepath
	}
	logger.Debug("temporary file created at", tmp.Name())
	defer tmp.Close()
	_, err = tmp.WriteString(decrypted)
	if err != nil {
		logger.Fatal("cannot write to the temporary file:", err)
		return filepath
	}
	// chmod 400 to the temporary file to prevent other users from reading it, that's the default permission for ssh private keys
	if err := tmp.Chmod(0o400); err != nil {
		logger.Fatal("cannot set the permission for the temporary file:", err)
		return filepath
	}

	return tmp.Name()
}

// cleanupTempFiles removes the temporary files created for decrypted private keys, it should be called after the ssh command is executed to ensure that the temporary files are removed even if the command fails
func cleanupTempFiles(args []string) {
	for _, arg := range args {
		if strings.Contains(path.Base(arg), ".inv-ssh-") {
			logger.Debug("removing the temporary file:", arg)
			if err := os.Remove(arg); err != nil {
				logger.Debug("cannot remove the temporary file:", err)
			}
		}
	}
}
