/*
 * Copyright (c) 2011, Dawid Ciężarkiewicz. All rights reserved.
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
 * MA 02110-1301  USA
 */

package gobtc

import (
	"bytes"
	"encoding/binary"
	"io"
)

const MAGIC_MAIN = 0xD9B4BEF9
const MAGIC_TESTNET = 0xDAB5BFFA

var versionCmdSig = [12]byte{'v', 'e', 'r', 's', 'i', 'o', 'n', 0, 0, 0, 0, 0}
var verackCmdSig = [12]byte{'v', 'e', 'r', 'a', 'c', 'k', 0, 0, 0, 0, 0, 0}

type NetAddr struct {
	Time    uint32
	Service uint64
	IP      [16]byte
	Port    uint16
}

type MsgHeader struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum uint32
}

type VersionCmdHeader struct {
	Version       int32
	Services      uint64
	Timestamp     int64
	AddrRecv      NetAddr
	AddrFrom      NetAddr
	Nonce         uint64
	SubversionNum string
	StartHeight   int32
}

func readEach(reader io.Reader, items []interface{}) error {
	for _, item := range items {
		if err := binary.Read(reader, binary.LittleEndian, item); err != nil {
			return err
		}
	}
	return nil
}

func writeEach(writer io.Writer, items []interface{}) error {
	for _, item := range items {
		if err := binary.Write(writer, binary.LittleEndian, item); err != nil {
			return err
		}
	}
	return nil
}

func read(reader io.Reader, item interface{}) error {
	return readEach(reader, []interface{}{item})
}

func write(writer io.Writer, item interface{}) error {
	return writeEach(writer, []interface{}{item})

}

func readNetAddr(reader io.Reader, addr *NetAddr, version int32, isVersionCmd bool) error {
	var err error

	if version >= 31402 && !isVersionCmd {
		err = read(reader, &addr.Time)

		if err != nil {
			return err
		}
	}
	err = readEach(reader, []interface{}{
		&addr.Service,
		&addr.IP,
		&addr.Port,
	})

	return err
}

func readVarInt(reader io.Reader, l *uint64) error {
	var err error
	var b byte

	err = read(reader, &b)

	if err != nil {
		return err
	}

	switch b {
	case 0xfd:
		var i uint16
		err = read(reader, &i)

		if err != nil {
			return err
		}

		*l = uint64(i)
	case 0xfe:
		var i uint32
		err = read(reader, &i)

		if err != nil {
			return err
		}

		*l = uint64(i)
	case 0xff:
		err = read(reader, l)

		if err != nil {
			return err
		}
	default:
		*l = uint64(b)
	}

	return nil
}

func readVarStr(reader io.Reader, str *string) error {
	var err error
	var l uint64

	err = readVarInt(reader, &l)

	if err != nil {
		return err
	}

	buf := make([]byte, l)

	if l > 0 {
		err = read(reader, &buf)
	}

	return err
}

func readVersionMsg(reader io.Reader, header *VersionCmdHeader) error {
	var err error

	err = readEach(reader, []interface{}{
		&header.Version,
		&header.Services,
		&header.Timestamp,
	})

	if err != nil {
		return err
	}

	err = readNetAddr(reader, &header.AddrRecv, header.Version, true)

	if header.Version < 106 {
		return err
	}

	err = readNetAddr(reader, &header.AddrFrom, header.Version, true)

	if err != nil {
		return err
	}

	err = readEach(reader, []interface{}{
		&header.Nonce,
	})

	if err != nil {
		return err
	}

	err = readVarStr(reader, &header.SubversionNum)

	if header.Version < 209 {
		return err
	}

	err = read(reader, &header.StartHeight)

	return err
}

func writeVerackMsg(writer io.Writer, magic uint32) error {
	return writeMsgHeader(
		writer,
		magic,
		&verackCmdSig,
		0,
	)
}

func readMsgHeader(reader io.Reader, header *MsgHeader) error {
	var err error

	err = readEach(reader, []interface{}{
		&header.Magic,
		&header.Command,
		&header.Length,
	})

	if err != nil {
		return err
	}

	if bytes.Compare(versionCmdSig[:], header.Command[:]) == 0 ||
		bytes.Compare(verackCmdSig[:], header.Command[:]) == 0 {
		return nil
	}

	err = read(reader, &header.Checksum)

	return err
}

func writeMsgHeader(writer io.Writer, magic uint32, cmdSig *[12]byte, l uint32) error {
	var err error

	err = write(writer, &struct {
		magic  uint32
		cmdSig [12]byte
		l      uint32
	}{
		magic,
		*cmdSig,
		l,
	})

	return err
}
