// this is a multiline comment
// that introduces the project.
//

// new value
owner: "<string>"

// counter value
counter: "<int>"

// should tls be enabled for the service 
enableTLS: true

// this is a required field with of type string with a constraint
redis_url: "https://redis.svc.url"

// gate is an endpoint that will be checked at the given interval 
gate: {
	// url 
	url: "test"

	// interval in miliseconds
	interval: 1000

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
			phone: "087 2853692"
		}
	}
}

// this is an optional field
labels: {}

// annotations can be added optionally
annotations: {}

// sets the replicas for the service
replicas: 2

// count is a series of numbers
// that have very little relevance to anyone
count: [1, 2, 3, 4, 5]

// replicas max
max: replicas * 3
