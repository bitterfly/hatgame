module Login exposing (Data, default)


type alias Data =
    { email : Maybe String
    , password : Maybe String
    }


default : Data
default =
    { email = Nothing, password = Nothing }
