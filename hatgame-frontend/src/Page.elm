module Page exposing (Page(..))

import Containers.Game
import Containers.User
import Home
import Host
import Login
import Started
import UserCredentials
import Words


type Page
    = Login Login.Data
    | Register UserCredentials.Data
    | Change Home.Data UserCredentials.Data
    | Home Home.Data
    | Host Host.Data
    | Lobby Containers.Game.Game
    | Words Words.Data
    | Started Started.Data
    | Ended (List Containers.User.User) (List Containers.Game.Team)
