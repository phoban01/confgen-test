// this is a multiline comment
// that introduces the project.
//

// sets the replicas for the service
replicas: 2

// replicas max
max: replicas * 3

// new value
owner: "<string>"

// counter value
counter: "<int>"

// should tls be enabled for the service 
enableTLS: true

// this is a required field with of type string with a constraint
redis_url: "<string>"

// this is an optional field
labels: {}

// annotations can be added optionally
annotations: {}

// gate is an endpoint that will be checked at the given interval 
gate: {
	// url 
	url: "test"

	// interval in miliseconds
	interval: _|_ // conflicting values 1000 and 10

	// data
	data: {
		// forename
		forname: "piaras"
		// surname
		surname: "<string>"

		// address
		address: {
			// home place
			home: "<string>"

			// phone number
			phone: "<string>"
		}
	}
}

// count is a series of numbers
// that have very little relevance to anyone
count: [1, 2, 3, 4, 5]
