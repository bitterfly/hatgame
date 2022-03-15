module Page exposing (Page(..))

import Home
import Host
import Lobby
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
    | Lobby Lobby.Data
    | Words Words.Data
    | Started Started.Data