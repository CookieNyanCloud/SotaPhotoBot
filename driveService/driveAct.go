package driveService

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"

	"google.golang.org/api/drive/v3"
)

type DriveService struct {
	Srv    *drive.Service
	People string
	Zag    string
}

func NewDriveService(srv *drive.Service) IDrive {
	return &DriveService{Srv: srv}
}

type IDrive interface {
	GetPhotos(name string) ([]*http.Response, error)
	SendPhotos(name, folder string, file *http.Response) error
	FindPhoto(name string) (*drive.FileList, error)
	Load(r *drive.File, wg *sync.WaitGroup) (*http.Response, error)
}

func (srv *DriveService) GetPhotos(name string) ([]*http.Response, error) {

	if len(name) < 7 {
		log.Printf("too short name: %v\n", len(name))
		return nil, errors.New("too short name")
	}

	r, err := srv.FindPhoto(name)
	if err != nil {
		log.Printf("Unable to retrieve files: %v\n", err)
		return nil, errors.New("unable to retrieve files")
	}
	fileslist := make([]*http.Response, len(r.Files))
	if len(r.Files) == 0 {
		log.Println("No files found")
		return nil, errors.New("no files found")
	} else {
		var wg sync.WaitGroup
		for i, driveFile := range r.Files {
			wg.Add(1)
			go func(srv *DriveService, driveFile *drive.File, wg *sync.WaitGroup, i int) {
				fileName, err := srv.Load(driveFile, wg)
				if err != nil {
					log.Printf("unable to retrieve files: %v\n", err)
				}
				fileslist[i] = fileName
			}(srv, driveFile, &wg, i)
		}
		wg.Wait()
	}
	return fileslist, nil
}

func (srv *DriveService) SendPhotos(name, folder string, file *http.Response) error {
	r, err := srv.FindPhoto(name)
	if err != nil {
		log.Printf("Unable to retrieve files: %v\n", err)
		return err
	}
	num := strconv.Itoa(len(r.Files) + 1)
	name += " " + num + ".jpeg"

	fl := &drive.File{}
	fl.Name = name
	fl.MimeType = "image/jpeg"
	fs := make([]string, 1)
	fs[0] = folder
	fl.Parents = fs
	fileDone, err := srv.Srv.
		Files.
		Create(fl).
		SupportsAllDrives(true).
		SupportsTeamDrives(true).
		Media(file.Body).
		Do()
	file.Body.Close()
	log.Println(fileDone.Name)
	return nil
}

func (srv *DriveService) FindPhoto(name string) (*drive.FileList, error) {
	query := `name contains '` + name + "'"
	r, err := srv.Srv.Files.
		List().
		PageSize(20).
		Fields("nextPageToken, files(id, name, parents, driveId)").
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		Q(query).
		IncludePermissionsForView("published").
		Do()
	return r, err
}

func (srv *DriveService) Load(r *drive.File, wg *sync.WaitGroup) (*http.Response, error) {
	log.Println(r.Name, "start")
	res, err := srv.Srv.Files.Get(r.Id).Download()
	if err != nil {
		return nil, err
	}
	wg.Done()
	log.Println(r.Name, "end")
	return res, nil
}
