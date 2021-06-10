<html>
    <head>
    </head>
    <body>
        <form action="/upload" method="POST" enctype="multipart/form-data">
            <label for="auth">auth:</label><br>
            <input type="password" id="auth" name="auth">
            <label for="upload">data:</label><br>
            <input type="file" id="upload" name="upload">
            <input type="submit" value="Submit">
        </form>
    </body>
</html>