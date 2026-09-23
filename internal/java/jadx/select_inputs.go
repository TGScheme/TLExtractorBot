package jadx

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Laky-64/gologging"
)

type dexFile struct {
	name string
	data []byte
}

func selectInputs(apk, dexDir, pkg string) []string {
	inputs, err := extractDexInputs(apk, dexDir, "L"+strings.ReplaceAll(pkg, ".", "/")+"/")
	if err != nil {
		gologging.Warn("jadx: decompiling the whole apk:", err)
		return []string{apk}
	}
	return inputs
}

func extractDexInputs(apk, dexDir, prefix string) ([]string, error) {
	dexes, err := readDexes(apk)
	if err != nil {
		return nil, err
	}
	if err = os.RemoveAll(dexDir); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(dexDir, os.ModePerm); err != nil {
		return nil, err
	}
	var inputs, names []string
	for _, dex := range dexes {
		defined, err := countClasses(dex.data, prefix)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", dex.name, err)
		}
		if defined == 0 {
			continue
		}
		target := path.Join(dexDir, dex.name)
		if err = os.WriteFile(target, dex.data, 0o644); err != nil {
			return nil, err
		}
		inputs = append(inputs, target)
		names = append(names, fmt.Sprintf("%s (%d classes)", dex.name, defined))
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no dex defines %s", prefix)
	}
	gologging.Info(fmt.Sprintf("jadx: decompiling %s out of %d dex files", strings.Join(names, ", "), len(dexes)))
	return inputs, nil
}

func readDexes(apk string) ([]dexFile, error) {
	archive, err := zip.OpenReader(apk)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = archive.Close()
	}()
	var dexes []dexFile
	for _, file := range archive.File {
		if path.Dir(file.Name) != "." || !strings.HasPrefix(file.Name, "classes") || !strings.HasSuffix(file.Name, ".dex") {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return nil, err
		}
		dexes = append(dexes, dexFile{name: file.Name, data: data})
	}
	if len(dexes) == 0 {
		return nil, errors.New("no dex file in the apk")
	}
	return dexes, nil
}

func countClasses(data []byte, prefix string) (int, error) {
	if len(data) < 0x70 || !bytes.HasPrefix(data, []byte("dex\n")) {
		return 0, errors.New("invalid dex header")
	}
	var err error
	u32 := func(offset uint64) uint32 {
		if err != nil || offset+4 > uint64(len(data)) {
			err = errors.New("offset out of bounds")
			return 0
		}
		return binary.LittleEndian.Uint32(data[offset:])
	}
	stringIDs, typeCount, typeIDs := uint64(u32(0x3c)), u32(0x40), uint64(u32(0x44))
	classCount, classDefs := u32(0x60), uint64(u32(0x64))
	defined := 0
	for i := range uint64(classCount) {
		typeIdx := u32(classDefs + i*32)
		if typeIdx >= typeCount {
			return 0, errors.New("type index out of bounds")
		}
		offset := uint64(u32(stringIDs + uint64(u32(typeIDs+uint64(typeIdx)*4))*4))
		if err != nil {
			return 0, err
		}
		for offset < uint64(len(data)) && data[offset]&0x80 != 0 {
			offset++
		}
		offset++
		if offset+uint64(len(prefix)) > uint64(len(data)) {
			return 0, errors.New("string offset out of bounds")
		}
		if bytes.HasPrefix(data[offset:], []byte(prefix)) {
			defined++
		}
	}
	return defined, err
}
