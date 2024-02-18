## REST API Documentation

The REST API is mainly used for authentication and very few things.
For other part, we are using the GraphQL API. Check the [GraphQL API Documentation](https://github.com/swiftwave-org/swiftwave/blob/develop/docs/api_docs.md) for more information.

---

### Authentication API
**POST** /auth/login

**Form Data**

| Key | Value Type | Example Value |
| --- |------------|---------------|
| username | string     | admin         |
| password | string     | 12345         |

**Example Response**

**200 OK**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDI2NjI5MzUsImlhdCI6MTcwMjY1OTMzNSwibmJmIjoxNzAyNjU5MzM1LCJ1c2VybmFtZSI6InRhbm1veXNydCJ9.5X9n8iEQy7UNcGfReH2Ap2WiSXZfFkQ0WJURMIyl_O0"
}
```

**400 Bad Request**
- Invalid username
  ```json
  {
    "message": "user does not exist"
  }
  ```
- Missing password
  ```json
  {
    "message": "incorrect password"
  }
  ```

---

### Upload Source Code API
**POST** /upload/code

**Form Data**

| Key | Value Type            | Example Value |
| --- |-----------------------|---------------|
| file | file (only tar files) | code.tar         |

**Example Response**

**200 OK**
```json
{
  "file": "d396973f-82a2-4e42-9273-404d9e4a6696.tar",
  "message": "file uploaded successfully"
}
```

**400 Bad Request**
- Missing file
  ```json
  {
    "message": "file not found"
  }
  ```
- Invalid file format
  ```json
  {
    "message": "file is not a tar file"
  }
  ```

---

### Download Backup API

**GET** /persistent-volume/backup/<backup_id>/download

**Example Response**

**200 OK**
- The file will be downloaded

**Any other status code**
- The file will not be downloaded

---

### Backup Filename API

**GET** /persistent-volume/backup/:id/filename

**Example Response**

**200 OK**
- Filename will be sent as response text

**Any other status code**
- The filename will be failed to send

---

### Upload File for Restore API
**POST** /persistent-volume/restore/<restore_id>/upload

**Form Data**

| Key | Value Type            | Example Value |
| --- |-----------------------|---------------|
| file | file (only tar files) | file.tar.gz   |

**Example Response**

**200 OK**
```json
{
  "message": "file uploaded successfully, you can now start the restore process",
}
```

**Any other status code**
- The file upload failed