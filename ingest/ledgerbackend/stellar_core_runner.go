package ledgerbackend

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type hcnetCoreRunnerInterface interface {
	catchup(from, to uint32) error
	runFrom(from uint32, hash string) error
	getMetaPipe() io.Reader
	getProcessExitChan() <-chan error
	close() error
}

type hcnetCoreRunner struct {
	executablePath    string
	configPath        string
	networkPassphrase string
	historyURLs       []string

	started  bool
	wg       sync.WaitGroup
	shutdown chan struct{}

	cmd *exec.Cmd
	// processExit channel receives an error when the process exited with an error
	// or nil if the process exited without an error.
	processExit chan error
	metaPipe    io.Reader
	tempDir     string
	nonce       string
}

func newHcnetCoreRunner(executablePath, configPath, networkPassphrase string, historyURLs []string) (*hcnetCoreRunner, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create temp dir
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("captive-hcnet-core-%x", random.Uint64()))
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "error creating subprocess tmpdir")
	}

	runner := &hcnetCoreRunner{
		executablePath:    executablePath,
		configPath:        configPath,
		networkPassphrase: networkPassphrase,
		historyURLs:       historyURLs,
		shutdown:          make(chan struct{}),
		processExit:       make(chan error),
		tempDir:           tempDir,
		nonce:             fmt.Sprintf("captive-hcnet-core-%x", r.Uint64()),
	}

	if configPath == "" {
		err := runner.writeConf()
		if err != nil {
			return nil, errors.Wrap(err, "error writing configuration")
		}
	}

	return runner, nil
}

func (r *hcnetCoreRunner) generateConfig() string {
	lines := []string{
		"# Generated file -- do not edit",
		"RUN_STANDALONE=true",
		"NODE_IS_VALIDATOR=false",
		"DISABLE_XDR_FSYNC=true",
		"UNSAFE_QUORUM=true",
		fmt.Sprintf(`NETWORK_PASSPHRASE="%s"`, r.networkPassphrase),
		fmt.Sprintf(`BUCKET_DIR_PATH="%s"`, filepath.Join(r.tempDir, "buckets")),
	}
	for i, val := range r.historyURLs {
		lines = append(lines, fmt.Sprintf("[HISTORY.h%d]", i))
		lines = append(lines, fmt.Sprintf(`get="curl -sf %s/{0} -o {1}"`, val))
	}
	// Add a fictional quorum -- necessary to convince core to start up;
	// but not used at all for our purposes. Pubkey here is just random.
	lines = append(lines,
		"[QUORUM_SET]",
		"THRESHOLD_PERCENT=100",
		`VALIDATORS=["GCZBOIAY4HLKAJVNJORXZOZRAY2BJDBZHKPBHZCRAIUR5IHC2UHBGCQR"]`)
	return strings.ReplaceAll(strings.Join(lines, "\n"), "\\", "\\\\")
}

func (r *hcnetCoreRunner) getConfFileName() string {
	if r.configPath != "" {
		return r.configPath
	}
	return filepath.Join(r.tempDir, "hcnet-core.conf")
}

func (*hcnetCoreRunner) getLogLineWriter() io.Writer {
	r, w := io.Pipe()
	br := bufio.NewReader(r)
	// Strip timestamps from log lines from captive hcnet-core. We emit our own.
	dateRx := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3} `)
	go func() {
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				break
			}
			line = dateRx.ReplaceAllString(line, "")
			fmt.Print(line)
		}
	}()
	return w
}

// Makes the temp directory and writes the config file to it; called by the
// platform-specific captiveHcnetCore.Start() methods.
func (r *hcnetCoreRunner) writeConf() error {
	conf := r.generateConfig()
	return ioutil.WriteFile(r.getConfFileName(), []byte(conf), 0644)
}

func (r *hcnetCoreRunner) createCmd(params ...string) (*exec.Cmd, error) {
	allParams := append([]string{"--conf", r.getConfFileName()}, params...)
	cmd := exec.Command(r.executablePath, allParams...)
	cmd.Dir = r.tempDir
	cmd.Stdout = r.getLogLineWriter()
	cmd.Stderr = r.getLogLineWriter()
	return cmd, nil
}

func (r *hcnetCoreRunner) runCmd(params ...string) error {
	cmd, err := r.createCmd(params...)
	if err != nil {
		return errors.Wrapf(err, "could not create `hcnet-core %v` cmd", params)
	}

	if err = cmd.Start(); err != nil {
		return errors.Wrapf(err, "could not start `hcnet-core %v` cmd", params)
	}

	if err = cmd.Wait(); err != nil {
		return errors.Wrapf(err, "error waiting for `hcnet-core %v` subprocess", params)
	}
	return nil
}

func (r *hcnetCoreRunner) catchup(from, to uint32) error {
	if r.started {
		return errors.New("runner already started")
	}
	if err := r.runCmd("new-db"); err != nil {
		return errors.Wrap(err, "error waiting for `hcnet-core new-db` subprocess")
	}

	rangeArg := fmt.Sprintf("%d/%d", to, to-from+1)
	cmd, err := r.createCmd(
		"catchup", rangeArg,
		"--metadata-output-stream", r.getPipeName(),
		"--replay-in-memory",
	)
	if err != nil {
		return errors.Wrap(err, "error creating `hcnet-core catchup` subprocess")
	}
	r.cmd = cmd
	r.metaPipe, err = r.start()
	if err != nil {
		return errors.Wrap(err, "error starting `hcnet-core catchup` subprocess")
	}
	r.started = true

	// Do not remove bufio.Reader wrapping. Turns out that each read from a pipe
	// adds an overhead time so it's better to preload data to a buffer.
	r.metaPipe = bufio.NewReaderSize(r.metaPipe, 1024*1024)
	return nil
}

func (r *hcnetCoreRunner) runFrom(from uint32, hash string) error {
	if r.started {
		return errors.New("runner already started")
	}
	var err error
	r.cmd, err = r.createCmd(
		"run",
		"--in-memory",
		"--start-at-ledger", fmt.Sprintf("%d", from),
		"--start-at-hash", hash,
		"--metadata-output-stream", r.getPipeName(),
	)
	if err != nil {
		return errors.Wrap(err, "error creating `hcnet-core run` subprocess")
	}
	r.metaPipe, err = r.start()
	if err != nil {
		return errors.Wrap(err, "error starting `hcnet-core run` subprocess")
	}
	r.started = true

	// Do not remove bufio.Reader wrapping. Turns out that each read from a pipe
	// adds an overhead time so it's better to preload data to a buffer.
	r.metaPipe = bufio.NewReaderSize(r.metaPipe, 1024*1024)
	return nil
}

func (r *hcnetCoreRunner) getMetaPipe() io.Reader {
	return r.metaPipe
}

func (r *hcnetCoreRunner) getProcessExitChan() <-chan error {
	return r.processExit
}

func (r *hcnetCoreRunner) close() error {
	var err1, err2 error

	if r.processIsAlive() {
		err1 = r.cmd.Process.Kill()
		r.cmd.Wait()
		r.cmd = nil
	}
	err2 = os.RemoveAll(r.tempDir)
	r.tempDir = ""

	if r.started {
		close(r.shutdown)
		r.wg.Wait()
		close(r.processExit)
	}
	r.started = false

	if err1 != nil {
		return errors.Wrap(err1, "error killing subprocess")
	}
	if err2 != nil {
		return errors.Wrap(err2, "error removing subprocess tmpdir")
	}
	return nil
}
