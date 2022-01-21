module UserCredentials exposing (Data, default)


type alias Data =
    { email : Maybe String
    , password : Maybe String
    , username : Maybe String
    }


default : Data
default =
    { email = Nothing, password = Nothing, username = Nothing }
