package main

import (
	"flag"
	"log"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
</head>
<body>
	<h1>Preflight CORS</h1>
	<div id="output"></div>
	<script>
		document.addEventListener('DOMContentLoaded',function(){
			fetch("http://localhost:3030/v1/tokens/authentication",{
				method: "POST",
				headers: {
					'Content-Type':'application/json'
				},
				body: JSON.stringify({
					email: 'sukratiagrawal192@gmail.com',
					password: 'sukratia'
				})
			}).then(
				function(response){
					response.text().then(function(text){
						document.getElementById("output").innerHTML=text;
					});
				},
				function(error){
					document.getElemetById("output").innerHTML=err;
				}
			);
		});
	</script>
</body>
</html>`

func main() {
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()
	log.Println("starting server on: ", *addr)
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
