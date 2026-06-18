/*
Package converter provides data transformation utilities for the VIDI middleware.

It handles conversion of form data and stored records into various output formats:

 1. JSON: Pretty-printed JSON objects and arrays
 2. XML: Simple XML wrapper with fields as elements
 3. SQL: Generated INSERT statements for PostgreSQL compatibility

# Usage

	import "github.com/Emperor42/vidi/pkg/converter"

	func main() {
		converter := converter.New()

		data := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}

		// Convert to JSON
		jsonBytes, _ := converter.ToJSON(data)

		// Convert to XML
		xmlBytes, _ := converter.ToXML(data)

		// Convert to SQL
		sqlBytes, _ := converter.ToSQL(data, "users")
	}

# Format Details

JSON:

	{
	  "name": "John Doe",
	  "email": "john@example.com"
	}

XML:

	<?xml version="1.0" encoding="UTF-8"?>
	<root>
	  <name>John Doe</name>
	  <email>john@example.com</email>
	</root>

SQL:

	INSERT INTO users (name, email) VALUES ($1, $2);

# Thread Safety

The Converter is stateless and safe for concurrent use.

# See Also

 1. VIDI package: github.com/Emperor42/vidi/pkg/vidi
 2. Storage package: github.com/Emperor42/vidi/pkg/storage

# License

MIT License - See LICENSE file for details.

# Author

Emperor42
*/
package converter
