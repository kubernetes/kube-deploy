package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golang/glog"
	"hash"
	"io"
	"os"
	"path"
	"strconv"
)

func WriteFile(destPath string, contents []byte, fileMode os.FileMode, dirMode os.FileMode) (bool, error) {
	changed := false

	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return changed, fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	var hash string
	{
		hasher := sha256.New()
		hasher.Write(contents)
		hash = hex.EncodeToString(hasher.Sum(nil))
	}

	hasHash, err := fileHasHash(destPath, hash)
	if err != nil {
		return changed, fmt.Errorf("error checking hash of file %q: %v", destPath, err)
	}

	if !hasHash {
		in := bytes.NewBuffer(contents)
		err = writeFileContentsAlways(destPath, in, fileMode)
		if err != nil {
			return changed, err
		}
		changed = true
	}

	modeChanged, err := ensureFileMode(destPath, fileMode)
	if err != nil {
		return false, err
	}
	changed = changed || modeChanged

	return changed, nil
}

func writeFileContentsAlways(destPath string, in io.Reader, fileMode os.FileMode) error {
	glog.Infof("Writing file %q", destPath)

	out, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("error opening destination file %q: %v", destPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", destPath, err)
	}
	return nil
}

func ensureFileMode(destPath string, fileMode os.FileMode) (bool, error) {
	changed := false
	stat, err := os.Stat(destPath)
	if err != nil {
		return changed, fmt.Errorf("error getting file mode for %q: %v", destPath, err)
	}
	if stat.Mode() == fileMode {
		return changed, nil
	}
	glog.Infof("Changing file mode for %q to %s", destPath, fileMode)

	err = os.Chmod(destPath, fileMode)
	if err != nil {
		return changed, fmt.Errorf("error setting file mode for %q: %v", destPath, err)
	}
	changed = true
	return changed, nil
}

func CopyFileIfDifferent(destPath string, srcPath string, fileMode os.FileMode, dirMode os.FileMode) (bool, error) {
	changed := false

	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return changed, fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	hash, err := HashFile(srcPath, sha256.New())
	if err != nil {
		return changed, fmt.Errorf("error hashing source file %q: %v", srcPath, err)
	}

	sameHash, err := fileHasHash(destPath, hash)
	if err != nil {
		return changed, fmt.Errorf("error checking hash for file %q: %v", destPath, err)
	}

	if !sameHash {
		in, err := os.OpenFile(srcPath, os.O_RDONLY, fileMode)
		if err != nil {
			return changed, fmt.Errorf("error opening source file %q: %v", srcPath, err)
		}
		defer in.Close()

		err = writeFileContentsAlways(destPath, in, fileMode)
		if err != nil {
			return changed, err
		}
		changed = true
	}

	// TODO: Only if file not created
	modeChanged, err := ensureFileMode(destPath, fileMode)
	if err != nil {
		return changed, err
	}
	changed = changed || modeChanged

	return changed, nil
}

func fileHasHash(f string, expected string) (bool, error) {
	var hasher hash.Hash
	if len(expected) == 32 {
		hasher = md5.New()
	} else if len(expected) == 40 {
		hasher = sha1.New()
	} else if len(expected) == 64 {
		hasher = sha256.New()
	} else {
		return false, fmt.Errorf("Unrecognized hash format: %q", expected)
	}

	actual, err := HashFile(f, hasher)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if actual == expected {
		glog.V(2).Infof("Hash matched for %q: %v", f, expected)
		return true, nil
	} else {
		glog.V(2).Infof("Hash did not match for %q: actual=%v vs expected=%v", f, actual, expected)
		return false, nil
	}
}

func HashFile(f string, hasher hash.Hash) (string, error) {
	glog.V(2).Infof("hashing file %q", f)
	in, err := os.OpenFile(f, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer in.Close()

	_, err = io.Copy(hasher, in)
	if err != nil {
		return "", fmt.Errorf("error hashing file %q: %v", f, err)
	}

	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func ParseFileMode(s string, defaultMode os.FileMode) (os.FileMode, error) {
	fileMode := defaultMode
	if s != "" {
		v, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return fileMode, fmt.Errorf("cannot parse file mode %q", s)
		}
		fileMode = os.FileMode(v)
	}
	return fileMode, nil
}
