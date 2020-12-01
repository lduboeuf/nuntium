/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of telepathy.
 *
 * mms is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * mms is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/ubports/nuntium/mms"
	"launchpad.net/go-xdg/v0"
)

const SUBPATH = "nuntium/store"

func Create(uuid, contentLocation string) error {
	state := MMSState{
		State:           NOTIFICATION,
		ContentLocation: contentLocation,
	}
	storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func Destroy(uuid string) error {
	log.Printf("storage.Destroy(%v)", uuid)
	if storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".db")); err == nil {
		if err := os.Remove(storePath); err != nil {
			return err
		}
	} else {
		return err
	}

	if mNotificationIndPath, err := GetMNotificationInd(uuid); err == nil {
		log.Printf("storage.Destroy: found mNotificationIndPath: %v", mNotificationIndPath)
		if err := os.Remove(mNotificationIndPath); err != nil {
			return err
		}
	} else {
		if mmsPath, err := GetMMS(uuid); err == nil {
			if err := os.Remove(mmsPath); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func CreateResponseFile(uuid string) (*os.File, error) {
	filePath, err := xdg.Cache.Ensure(path.Join(SUBPATH, uuid+".m-notifyresp.ind"))
	if err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

func UpdateFailed(mNotificationInd *mms.MNotificationInd) error {
	log.Printf("storage.UpdateFailed(%v)", mNotificationInd)
	if mmsPath, err := GetMMS(mNotificationInd.UUID); err == nil {
		log.Printf("storage.UpdateFailed: found %v, deleting", mmsPath)
		if err := os.Remove(mmsPath); err != nil {
			return err
		}
	}
	//TODO:jezek Create this file on Create()? Do we really need it? Isn't ContentLocation enough?
	log.Printf("storage.UpdateFailed: saving mNotificationInd to file %v: %#v", path.Join(SUBPATH, mNotificationInd.UUID+".m-notify.ind"), mNotificationInd)
	mNotificationIndPath, err := xdg.Data.Ensure(path.Join(SUBPATH, mNotificationInd.UUID+".m-notify.ind"))
	if err != nil {
		return err
	}

	file, err := os.Create(mNotificationIndPath)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(mNotificationIndPath)
		}
	}()

	w := bufio.NewWriter(file)
	defer w.Flush()
	jsonWriter := json.NewEncoder(w)
	if err := jsonWriter.Encode(mNotificationInd); err != nil {
		return err
	}
	return nil
}

func UpdateDownloaded(uuid, filePath string) error {
	mmsPath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".mms"))
	if err != nil {
		return err
	}
	log.Printf("storage.UpdateDownloaded(%v): rename %v to %v", uuid, filePath, mmsPath)
	if err := os.Rename(filePath, mmsPath); err != nil {
		os.Remove(mmsPath)
		return err
	}

	if mNotificationIndPath, err := GetMNotificationInd(uuid); err == nil {
		log.Printf("storage.UpdateDownloaded(%v): found mNotificationIndPath: %v", uuid, mNotificationIndPath)
		if err := os.Remove(mNotificationIndPath); err != nil {
			return err
		}
	}

	state := MMSState{
		State: DOWNLOADED,
	}
	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func UpdateRetrieved(uuid string) error {
	state := MMSState{
		State: RETRIEVED,
	}
	storePath, err := xdg.Data.Find(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return err
	}
	return writeState(state, storePath)
}

func CreateSendFile(uuid string) (*os.File, error) {
	state := MMSState{
		State: DRAFT,
	}
	storePath, err := xdg.Data.Ensure(path.Join(SUBPATH, uuid+".db"))
	if err != nil {
		return nil, err
	}
	if err := writeState(state, storePath); err != nil {
		os.Remove(storePath)
		return nil, err
	}
	filePath, err := xdg.Cache.Ensure(path.Join(SUBPATH, uuid+".m-send.req"))
	if err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

func GetMMS(uuid string) (string, error) {
	return xdg.Data.Find(path.Join(SUBPATH, uuid+".mms"))
}

func GetMNotificationInd(uuid string) (string, error) {
	log.Printf("storage.GetMNotificationInd: looking for: %v", path.Join(SUBPATH, uuid+".m-notify.ind"))
	return xdg.Data.Find(path.Join(SUBPATH, uuid+".m-notify.ind"))
}

func writeState(state MMSState, storePath string) error {
	file, err := os.Create(storePath)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(storePath)
		}
	}()
	w := bufio.NewWriter(file)
	defer w.Flush()
	jsonWriter := json.NewEncoder(w)
	if err := jsonWriter.Encode(state); err != nil {
		return err
	}
	return nil
}
