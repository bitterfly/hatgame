module Containers.User exposing
    ( LoginUser
    , RegisteredUser
    , User
    , WithToken
    , decode
    , decodeUser
    , encodeLogin
    , encodeRegister
    , maybeUsername
    , maybeUsernameWithToken
    , show
    )

import Json.Decode exposing (Decoder)
import Json.Encode


type alias User =
    { id : Int, email : String, username : String }


type alias WithToken =
    { sessionToken : String, user : User }


maybeUsername : Maybe User -> String
maybeUsername user =
    case user of
        Nothing ->
            ""

        Just u ->
            u.username


maybeUsernameWithToken : Maybe WithToken -> String
maybeUsernameWithToken withToken =
    case withToken of
        Nothing ->
            ""

        Just u ->
            u.user.username


decodeUser : Decoder User
decodeUser =
    Json.Decode.map3 User
        (Json.Decode.field "ID" Json.Decode.int)
        (Json.Decode.field "Email" Json.Decode.string)
        (Json.Decode.field "Username" Json.Decode.string)


decode : Decoder WithToken
decode =
    Json.Decode.map2 WithToken
        (Json.Decode.field "sessionToken" Json.Decode.string)
        (Json.Decode.field "user" decodeUser)


show : User -> String
show { username } =
    "[" ++ username ++ "]"


type alias LoginUser =
    { email : String, password : String }


type alias RegisteredUser =
    { email : String, password : String, username : String }


encodeRegister : RegisteredUser -> Json.Encode.Value
encodeRegister { email, password, username } =
    Json.Encode.object
        [ ( "Email", Json.Encode.string email )
        , ( "Password", Json.Encode.string password )
        , ( "Username", Json.Encode.string username )
        ]


encodeLogin : LoginUser -> Json.Encode.Value
encodeLogin { email, password } =
    Json.Encode.object
        [ ( "Email", Json.Encode.string email )
        , ( "Password", Json.Encode.string password )
        ]
