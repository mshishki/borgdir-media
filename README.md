##### borgdir.media

Install following packages
`go get github.com/gorilla/sessions`
`go get -u golang.org/x/crypto/`
`go get github.com/mattn/go-sqlite3`

Type localhost:8080 in your browser, and you're in!

### Parts that are cool:
* `/equipment` page
![What every user can access](https://github.com/mshishki/borgdir-media/blob/master/Bildschirmfoto%202018-07-03%20um%2015.23.50.png "borgdir.media/equipment: Accessible to everyone")

* `/cart` and the ability to reserve items
![Accessible to logged-in users](https://github.com/mshishki/borgdir-media/blob/master/Bildschirmfoto%202018-07-03%20um%2015.28.38.png "borgdir.media/cart: Accessible to logged-in users")

* `/admin/...` resources
![Accessible to admins](https://github.com/mshishki/borgdir-media/blob/master/Bildschirmfoto%202018-07-03%20um%2015.32.18.png "borgdir.media/admin/edit-clients: Accessible to logged-in users")
(*Use admin/root as your username-password combination*)


## Parts that really, really, really suck and should've been altered, but will be ignored despite that, since the project was OK'd and I still have a life to live:

[ ] register datatypes for storage in sessions via gob/encoding

[ ] delete orders whenever Return is reached

[ ] change styling of individual Profile, Equipment etc. pages because this is horrible lol


# Things I could do but am too unsure of, because they might actually turn out to be counterproductive

[ ] http.Redirect only works with 3xx Status Codes, which is not always the case, but `w.WriteHeader(200)` right before redirecting with a 3xx is just so... redundant???

[ ] define methods for each page URL in **main.go** instead of switching between cases of an `*http.Request` instance

