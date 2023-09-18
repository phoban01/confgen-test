// this is a multiline comment
// that introduces the project.
//

// sets the replicas for the service
replicas: 2

// replicas max
max: replicas * 3

// new value
owner: "piaras.hoban@sap.com"

// counter value
counter: 100

// should tls be enabled for the service 
enableTLS: true

// this is a required field with of type string with a constraint
redis_url: "https://redis:6379"

// this is an optional field
labels: {}

// annotations can be added optionally
annotations: {}

// gate is an endpoint that will be checked at the given interval 
gate: {
	// url 
	url: "test"

	// data
	data: {
		// forename
		forname: "piaras"
		// surname
		surname: "hoban"

		// address
		address: {
			// home place
			home: "<string>"

			// phone number
			phone: "<string>"
		}
	}
}

// this is a new value
id: "transactions"

// count is a series of numbers
// that have very little relevance to anyone
count: [1, 2, 3, 4, 5]
