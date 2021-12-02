package driveService

import (
	"errors"
	"google.golang.org/api/drive/v3"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

type DriveService struct {
	Srv    *drive.Service
	People string
	Zag    string
}

func NewDriveService(srv *drive.Service, people, zag string) IDrive {
	return &DriveService{Srv: srv}
}

type IDrive interface {
	GetPhotos(name string) ([]os.File, error)
	SendPhotos(name, folder string, file os.File) error
	FindPhoto(name string) (*drive.FileList, error)
	Load(r *drive.File, wg *sync.WaitGroup, file *os.File) error
}

func (srv *DriveService) GetPhotos(name string) ([]os.File, error) {

	if len(name) < 7 {
		log.Printf("too short name: %v\n", len(name))
		return []os.File{}, errors.New("too short name")
	}

	r, err := srv.FindPhoto(name)
	if err != nil {
		log.Printf("Unable to retrieve files: %v\n", err)
		return []os.File{}, errors.New("unable to retrieve files")
	}
	fileslist := make([]os.File, len(r.Files))
	if len(r.Files) == 0 {
		log.Println("No files found")
		return []os.File{}, errors.New("no files found")
	} else {
		var wg sync.WaitGroup
		for j, i := range r.Files {
			wg.Add(1)
			go func(srv *DriveService, i *drive.File, wg *sync.WaitGroup, file *os.File) {
				err = srv.Load(i, wg, file)
				if err != nil {
					log.Printf("unable to retrieve files: %v\n", err)
				}
			}(srv, i, &wg, &fileslist[j])
		}
		wg.Wait()
	}
	return fileslist, nil
}

func (srv *DriveService) SendPhotos(name, folder string, file os.File) error {
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
		Media(&file).
		Do()
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

func (srv *DriveService) Load(r *drive.File, wg *sync.WaitGroup, file *os.File) error {
	log.Println(r.Name, "start")
	res, err := srv.Srv.Files.Get(r.Id).Download()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}
	wg.Done()
	log.Println(r.Name, "done")
	return nil
}
