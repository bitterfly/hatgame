module Msg exposing (..)

import Containers.Game
import Containers.Statistics
import Containers.User
import Home
import Host
import Http
import Login
import UserCredentials
import Words


type Field
    = Email
    | Password


type Msg
    = ChangeLogin Login.Data
    | ChangeRegister UserCredentials.Data
    | ChangeChangePage UserCredentials.Data
    | ChangeHost Host.Data
    | ChangeWords Words.Data
    | ChangeHome Home.Data
    | Login
    | ToRegisterPage
    | CheckRegister (Result Http.Error ())
    | ToHomePage (Result Http.Error ())
    | ToChangePage Home.Data
    | Register
    | ChangeUserData
    | GotUserToken (Result Http.Error Containers.User.WithToken)
    | GotCurrentUserToken (Result Http.Error Containers.User.WithToken)
    | GameOk (Result Http.Error ())
    | GotGame (Result Http.Error Containers.Game.Game)
    | GotStats (Result Http.Error Containers.Statistics.Statistics)
    | Create
    | Nothing
    | Host
    | Join Home.Data
    | CheckGame Home.Data
    | Rcv String
    | SendWord Words.Data
    | SendGuessed String
    | RemoveError
    | SendReady
    | End
