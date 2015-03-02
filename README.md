# How to Install
You must have Ruby and bundler installed.
First, download/Clone this repository, then

```bash
cd /path/to/vquery  
bundle install
```

# How to use

1. Log in to ESWeb, design and save your query.
2. Get the query id from the end of the URL  
  Example: "esweb.asp?WCI=Results&Query=139186"
  The query ID is 139186
3. run ```ruby vquery.rb QUERY_ID```  
  Example: ```ruby vquery.rb 139186```

# Credentials

This script only works with environment variables.  
VERACROSS_USERNAME  
VERACROSS_PASSWORD  
VERACROSS_CLIENT  
The client is the part of your URL that identifies your school.

# Why
The official API doesn't do a lot of things I want.
