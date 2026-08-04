package fbhttp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filebrowser/filebrowser/v2/runner"
	"github.com/filebrowser/filebrowser/v2/users"
)

const (
	WSWriteDeadline = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotAllowed = []byte("Command not allowed.")
)

func commandWorkingDirectory(user *users.User, path string) (string, error) {
	info, err := user.Fs.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("command working directory is not a directory")
	}

	return user.FullPath(path), nil
}

func wsErr(ws *websocket.Conn, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("command websocket failed: status=%d error=%T", status, err)
	}
	if err := ws.WriteControl(websocket.CloseInternalServerErr, []byte(txt), time.Now().Add(WSWriteDeadline)); err != nil {
		log.Print(err)
	}
}

var commandsHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer conn.Close()

	var raw string

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			wsErr(conn, http.StatusInternalServerError, err)
			return 0, nil
		}

		raw = strings.TrimSpace(string(msg))
		if raw != "" {
			break
		}
	}

	// Fail fast
	if !d.server.EnableExec || !d.user.Perm.Execute {
		if err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed); err != nil {
			wsErr(conn, http.StatusInternalServerError, err)
		}

		return 0, nil
	}

	command, err := runner.ParseAllowedCommand(raw, d.user.Commands)
	if err != nil {
		if errors.Is(err, runner.ErrCommandNotAllowed) {
			if err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed); err != nil {
				wsErr(conn, http.StatusInternalServerError, err)
			}
			return 0, nil
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			wsErr(conn, http.StatusInternalServerError, err)
		}
		return 0, nil
	}

	workingDirectory, err := commandWorkingDirectory(d.user, r.URL.Path)
	if err != nil {
		wsErr(conn, errToStatus(err), err)
		return 0, nil
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = workingDirectory

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		wsErr(conn, http.StatusInternalServerError, err)
		return 0, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		wsErr(conn, http.StatusInternalServerError, err)
		return 0, nil
	}

	if err := cmd.Start(); err != nil {
		wsErr(conn, http.StatusInternalServerError, err)
		return 0, nil
	}

	s := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for s.Scan() {
		if err := conn.WriteMessage(websocket.TextMessage, s.Bytes()); err != nil {
			log.Print(err)
		}
	}

	if err := cmd.Wait(); err != nil {
		wsErr(conn, http.StatusInternalServerError, err)
	}

	return 0, nil
})
