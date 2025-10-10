package core

import (
	"net/http"
	"strings"

	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/spheres-service/routes"
)

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()

	// spheres
	mux.HandleFunc("/spheres", ownhttp.WithLogging("CreateSphere", routes.CreateSphere(s.spheresColl)))

	// contents
	mux.HandleFunc("/spheres/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/spheres/")

		switch {

		case strings.Contains(path, "/media/presign"):
			ownhttp.WithLogging("PresignSphereMedia",
				routes.PresignSphereMedia(s.mediaColl, s.S3Cfg, s.Presigner),
			)(w, r)

		// mark media as completed
		case strings.Contains(path, "/media/complete") && r.Method == http.MethodPost:
			ownhttp.WithLogging("CompleteSphereMedia",
				routes.CompleteSphereMedia(s.mediaColl, s.S3Cfg, s.S3c),
			)(w, r)

		case strings.Contains(path, "/media/") && r.Method == http.MethodDelete:
			ownhttp.WithLogging("DeleteSphereMedia",
				routes.DeleteSphereMedia(s.mediaColl, s.S3Cfg, s.S3c),
			)(w, r)

		// create content
		case strings.HasSuffix(path, "/contents") && r.Method == http.MethodPost:
			ownhttp.WithLogging("CreateSphereContent",
				routes.CreateSphereContent(s.sphereContentsColl, s.mediaColl, s.EventStore),
			)(w, r)

			// list content (paginable, accept parentId y after)
		// list posts (paginable, with preview of replies )
		case strings.HasSuffix(path, "/contents") && r.Method == http.MethodGet:
			ownhttp.WithLogging("GetSpherePosts",
				routes.GetSpherePosts(s.sphereContentsColl),
			)(w, r)

		// list replies of a post
		case strings.Contains(path, "/contents/") && strings.HasSuffix(path, "/replies") && r.Method == http.MethodGet:
			ownhttp.WithLogging("GetSphereReplies",
				routes.GetSphereReplies(s.sphereContentsColl),
			)(w, r)

		// update content
		case strings.Contains(path, "/contents/") && r.Method == http.MethodPatch:
			ownhttp.WithLogging("UpdateSphereContent",
				routes.UpdateSphereContent(s.sphereContentsColl, s.EventStore),
			)(w, r)

		// remove content
		case strings.Contains(path, "/contents/") && r.Method == http.MethodDelete && !strings.HasSuffix(path, "/reactions"):
			ownhttp.WithLogging("DeleteSphereContent",
				routes.DeleteSphereContent(s.sphereContentsColl, s.EventStore),
			)(w, r)

		// reactions (add/remove)
		case strings.Contains(path, "/contents/") && strings.HasSuffix(path, "/reactions"):
			ownhttp.WithLogging("ReactSphereContent",
				routes.ReactSphereContent(s.sphereContentsColl, s.EventStore),
			)(w, r)

		default:
			ownhttp.WriteJSONError(w, 404, "NOT_FOUND", "route")
		}
	})

	return mux
}
