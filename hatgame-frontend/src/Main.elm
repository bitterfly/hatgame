port module Main exposing (main)

import Browser
import Browser.Navigation exposing (Key)
import Change.Http
import Change.View
import Containers.Game
import Containers.Message
import Containers.User
import Ended.View
import Generic.Utils
import Home.Http
import Home.View
import Host
import Host.View
import Html exposing (Html, br, div, text)
import Html.Attributes exposing (..)
import Http
import Json.Decode
import Lobby
import Lobby.View
import Login
import Login.Http
import Login.View
import Msg exposing (Msg)
import Page exposing (Page(..))
import Process
import Register.Http
import Register.View
import Started exposing (ProcessState(..))
import Started.View
import Task
import Time
import Url exposing (Url)
import Words
import Words.View



-- MAIN


main : Program String Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlRequest = onUrlRequest
        , onUrlChange = onUrlChange
        }



-- application :
--     { init : flags -> Url -> Key -> ( model, Cmd msg )
--     , view : model -> Document msg
--     , update : msg -> model -> ( model, Cmd msg )
--     , subscriptions : model -> Sub msg
--     , onUrlRequest : UrlRequest -> msg
--     , onUrlChange : Url -> msg
--     }
-- element :
--     { init : flags -> ( model, Cmd msg )
--     , view : model -> Html msg
--     , update : msg -> model -> ( model, Cmd msg )
--     , subscriptions : model -> Sub msg
--     }


onUrlRequest : Browser.UrlRequest -> Msg
onUrlRequest _ =
    Msg.Nothing


onUrlChange : Url -> Msg
onUrlChange _ =
    Msg.Nothing



-- MODEL


type alias Model =
    { page : Page.Page
    , tokenUser : Maybe Containers.User.WithToken
    , backend : String
    , error : Maybe String
    , key : Key
    }


init : String -> Url -> Key -> ( Model, Cmd Msg )
init backend _ key =
    ( { page = Page.Login Login.default
      , tokenUser = Nothing
      , backend = backend
      , error = Nothing
      , key = key
      }
    , Cmd.none
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Msg.ChangeLogin loginData ->
            ( case model.page of
                Page.Login _ ->
                    { model | page = Page.Login loginData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.Login ->
            ( model
            , handleLogin model
            )

        Msg.ChangeRegister registerData ->
            ( case model.page of
                Page.Register _ ->
                    { model | page = Page.Register registerData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.ToRegisterPage ->
            case model.page of
                Page.Login loginData ->
                    ( { model
                        | page =
                            Page.Register
                                { email = loginData.email
                                , password = Nothing
                                , username = Nothing
                                }
                      }
                    , Cmd.none
                    )

                _ ->
                    ( model, Cmd.none )

        Msg.Register ->
            case model.page of
                Page.Register { email, password, username } ->
                    ( model
                    , case ( email, password, username ) of
                        ( Just e, Just p, Just u ) ->
                            Register.Http.request model
                                { email = e
                                , password = p
                                , username = u
                                }

                        _ ->
                            Cmd.none
                    )

                _ ->
                    ( model
                    , Cmd.none
                    )

        Msg.ChangeUserData ->
            case model.page of
                Page.Change _ { email, password, username } ->
                    ( model
                    , case ( model.tokenUser, email ) of
                        ( Just t, Just e ) ->
                            Change.Http.request model
                                t
                                { email = e
                                , password = Maybe.withDefault "" password
                                , username = Maybe.withDefault "" username
                                }

                        _ ->
                            Cmd.none
                    )

                _ ->
                    ( model
                    , Cmd.none
                    )

        Msg.ChangeChangePage changeData ->
            ( case model.page of
                Page.Change homeData _ ->
                    { model | page = Page.Change homeData changeData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.ToChangePage homeData ->
            ( case model.tokenUser of
                Nothing ->
                    { model
                        | page =
                            Page.Change homeData
                                { email = Nothing
                                , password = Nothing
                                , username = Nothing
                                }
                    }

                Just t ->
                    { model
                        | page =
                            Page.Change homeData
                                { email = Just t.user.email
                                , password = Nothing
                                , username = Just t.user.username
                                }
                    }
            , Cmd.none
            )

        Msg.CheckRegister res ->
            case res of
                Err err ->
                    ( case err of
                        Http.BadStatus 409 ->
                            case model.page of
                                Page.Register registerData ->
                                    { model
                                        | page = Page.Register { registerData | email = Nothing }
                                        , error = Just "Email already in use."
                                    }

                                _ ->
                                    { model | error = Just <| Generic.Utils.errorToString err }

                        _ ->
                            { model
                                | page = Page.Login Login.default
                                , error = Just <| Generic.Utils.errorToString err
                            }
                    , hideError
                    )

                Ok _ ->
                    ( { model | page = Page.Login Login.default }, Cmd.none )

        Msg.ToHomePage res ->
            case model.page of
                Page.Change homeData _ ->
                    case res of
                        Err err ->
                            ( { model
                                | page = Page.Home homeData
                                , error = Just <| Generic.Utils.errorToString err
                              }
                            , hideError
                            )

                        Ok _ ->
                            ( { model | page = Page.Home homeData }
                            , case model.tokenUser of
                                Nothing ->
                                    Cmd.none

                                Just t ->
                                    Home.Http.getCurrentUserToken model t
                            )

                _ ->
                    ( { model
                        | page =
                            Page.Home
                                { gameId = Nothing
                                , stats = Nothing
                                }
                      }
                    , Cmd.none
                    )

        Msg.GotCurrentUserToken tokenUser ->
            case tokenUser of
                Err err ->
                    ( { model
                        | error = Just <| Generic.Utils.errorToString err
                      }
                    , hideError
                    )

                Ok t ->
                    ( { model
                        | tokenUser = Just t
                      }
                    , Home.Http.getStats model t
                    )

        Msg.GotUserToken tokenUser ->
            case tokenUser of
                Err err ->
                    ( { model
                        | error =
                            case err of
                                Http.BadStatus 401 ->
                                    Just "Wrong email or password."

                                _ ->
                                    Just <| Generic.Utils.errorToString err
                      }
                    , hideError
                    )

                Ok t ->
                    ( { model
                        | page =
                            Page.Home
                                { gameId = Nothing
                                , stats = Nothing
                                }
                        , tokenUser = Just t
                      }
                    , Home.Http.getStats model t
                    )

        Msg.Create ->
            ( { model | page = Page.Host Host.default }, Cmd.none )

        Msg.Nothing ->
            ( model, Cmd.none )

        Msg.ChangeHost hostData ->
            ( case model.page of
                Page.Host _ ->
                    { model | page = Page.Host hostData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.ChangeWords wordsData ->
            ( case model.page of
                Page.Words _ ->
                    { model | page = Page.Words wordsData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.ChangeHome homeData ->
            ( case model.page of
                Page.Home _ ->
                    { model | page = Page.Home homeData }

                _ ->
                    model
            , Cmd.none
            )

        Msg.Host ->
            case model.page of
                Page.Host hostData ->
                    case ( model.tokenUser, hostData.words, hostData.timer ) of
                        ( Just s, Just w, Just t ) ->
                            ( model
                            , sendHost
                                ( s.sessionToken
                                , [ hostData.players, w, hostData.stages, t ]
                                )
                            )

                        ( Just s, Nothing, Nothing ) ->
                            ( model
                            , sendHost
                                ( s.sessionToken
                                , [ hostData.players, Host.defaultWords, hostData.stages, Host.defaultTimer ]
                                )
                            )

                        _ ->
                            ( model, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        Msg.GameOk res ->
            case model.page of
                Page.Home homeData ->
                    case res of
                        Err err ->
                            ( { model
                                | page = Page.Home { homeData | gameId = Nothing }
                                , error =
                                    Just <|
                                        case err of
                                            Http.BadStatus 400 ->
                                                "The game doesn't exist."

                                            _ ->
                                                Generic.Utils.errorToString err
                              }
                            , hideError
                            )

                        Ok _ ->
                            case ( model.tokenUser, homeData.gameId ) of
                                ( Just s, Just n ) ->
                                    ( model
                                    , sendJoin ( s.sessionToken, n )
                                    )

                                _ ->
                                    ( model, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        Msg.GotGame res ->
            case res of
                Err err ->
                    ( { model
                        | error = Just <| Generic.Utils.errorToString err
                      }
                    , hideError
                    )

                Ok game ->
                    case model.tokenUser of
                        Nothing ->
                            ( model, Cmd.none )

                        Just s ->
                            ( { model
                                | page =
                                    Page.Lobby
                                        { game = game
                                        , processState = Lobby.WaitingPlayers
                                        }
                              }
                            , sendJoin ( s.sessionToken, game.id )
                            )

        Msg.GotStats res ->
            case res of
                Err err ->
                    ( { model
                        | error = Just <| Generic.Utils.errorToString err
                      }
                    , hideError
                    )

                Ok stats ->
                    case model.page of
                        Page.Home homeData ->
                            ( { model | page = Page.Home { homeData | stats = Just stats } }, Cmd.none )

                        _ ->
                            ( model, Cmd.none )

        Msg.CheckGame homeData ->
            case ( model.tokenUser, homeData.gameId ) of
                ( Just user, Just gameId ) ->
                    ( model, Home.Http.request model user { gameId = gameId } )

                _ ->
                    ( model, Cmd.none )

        Msg.Join homeData ->
            case ( model.tokenUser, homeData.gameId ) of
                ( Just s, Just n ) ->
                    ( model
                    , sendJoin ( s.sessionToken, n )
                    )

                _ ->
                    ( model, Cmd.none )

        Msg.Rcv m ->
            case Json.Decode.decodeString Containers.Message.decode m of
                Err err ->
                    ( { model
                        | error = Just <| Json.Decode.errorToString err
                      }
                    , hideError
                    )

                Ok (Containers.Message.Game game) ->
                    ( { model
                        | page =
                            Lobby
                                { game = game
                                , processState = Lobby.WaitingPlayers
                                }
                      }
                    , Cmd.none
                    )

                Ok Containers.Message.ReadyToStart ->
                    ( case model.page of
                        Lobby lobbyData ->
                            { model
                                | page =
                                    Lobby
                                        { game = lobbyData.game
                                        , processState = Lobby.ReadyToStart
                                        }
                            }

                        _ ->
                            model
                    , Cmd.none
                    )

                Ok (Containers.Message.ReceiveAddWord word) ->
                    case model.page of
                        Page.Words wordsData ->
                            let
                                newLen =
                                    List.length wordsData.words + 1
                            in
                            case compare newLen wordsData.game.numWords of
                                EQ ->
                                    ( { model
                                        | page =
                                            Words
                                                { game = wordsData.game
                                                , currentWord = Nothing
                                                , words = word :: wordsData.words
                                                , processState = Words.Done
                                                }
                                      }
                                    , Cmd.none
                                    )

                                LT ->
                                    ( { model
                                        | page =
                                            Words
                                                { wordsData
                                                    | currentWord = Nothing
                                                    , words = word :: wordsData.words
                                                }
                                      }
                                    , Cmd.none
                                    )

                                GT ->
                                    ( { model
                                        | page =
                                            Words
                                                { wordsData
                                                    | currentWord = Nothing
                                                }
                                      }
                                    , Cmd.none
                                    )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.Team partner) ->
                    case model.page of
                        Page.Words wordsData ->
                            ( { model
                                | page =
                                    Started
                                        { game = wordsData.game
                                        , currentWord = Nothing
                                        , partner = Containers.Game.playerById partner wordsData.game.players
                                        , timer = Nothing
                                        , processState = Started.NotStoryteller Nothing
                                        , results = []
                                        }
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.Tick timer) ->
                    case model.page of
                        Page.Started startedData ->
                            ( case startedData.processState of
                                Started.BetweenStages ->
                                    model

                                _ ->
                                    case timer of
                                        0 ->
                                            { model
                                                | page =
                                                    Started
                                                        { startedData
                                                            | currentWord = Nothing
                                                            , timer = Nothing
                                                        }
                                            }

                                        n ->
                                            { model | page = Started { startedData | timer = Just n } }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.Story word) ->
                    case model.page of
                        Page.Started startedData ->
                            ( { model
                                | page =
                                    Started
                                        { startedData
                                            | currentWord = Just word
                                            , processState = Started.StorytellerActive
                                        }
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.GuessPhaseStart id) ->
                    case model.page of
                        Page.Started startedData ->
                            case startedData.processState of
                                Started.BetweenStages ->
                                    ( model, Cmd.none )

                                _ ->
                                    case model.tokenUser of
                                        Nothing ->
                                            ( model, Cmd.none )

                                        Just t ->
                                            if t.user.id == id then
                                                ( { model
                                                    | page =
                                                        Started
                                                            { startedData
                                                                | processState = Started.StorytellerWaiting
                                                                , timer = Nothing
                                                            }
                                                  }
                                                , Cmd.none
                                                )

                                            else
                                                ( { model
                                                    | page =
                                                        Started
                                                            { startedData
                                                                | timer = Nothing
                                                                , processState =
                                                                    Started.NotStoryteller <|
                                                                        Containers.Game.playerById id startedData.game.players
                                                            }
                                                  }
                                                , Cmd.none
                                                )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.GameEnded results) ->
                    case model.page of
                        Page.Started startedData ->
                            ( { model
                                | page =
                                    Ended startedData.game.players <|
                                        Started.addResultsAndSort results startedData.results
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                Ok (Containers.Message.StageEnded results) ->
                    case model.page of
                        Page.Started startedData ->
                            ( { model
                                | page =
                                    Started
                                        { startedData
                                            | currentWord = Nothing
                                            , timer = Nothing
                                            , processState = Started.BetweenStages
                                            , results =
                                                Started.addResultsAndSort
                                                    results
                                                    startedData.results
                                        }
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                Ok Containers.Message.WordPhaseStart ->
                    ( case model.page of
                        Lobby lobbyData ->
                            { model
                                | page =
                                    Words
                                        { game = lobbyData.game
                                        , words = []
                                        , currentWord = Nothing
                                        , processState = Words.Typing
                                        }
                            }

                        _ ->
                            model
                    , Cmd.none
                    )

                Ok Containers.Message.ForcefullyEnded ->
                    ( { model | page = Home { gameId = Nothing, stats = Nothing } }
                    , case model.tokenUser of
                        Nothing ->
                            Cmd.none

                        Just t ->
                            Home.Http.getStats model t
                    )

                Ok (Containers.Message.Error err) ->
                    ( { model
                        | page =
                            case model.page of
                                Page.Home homeData ->
                                    Page.Home { gameId = Nothing, stats = homeData.stats }

                                _ ->
                                    model.page
                        , error = Just err
                      }
                    , hideError
                    )

        Msg.SendQuitLobby ->
            ( model
            , sendMessage <|
                Containers.Message.encodeMsgSend Containers.Message.SendQuitLobby
            )

        Msg.SendWord wordsData ->
            case wordsData.currentWord of
                Just w ->
                    ( { model
                        | page =
                            Words
                                { wordsData
                                    | currentWord = Nothing
                                }
                      }
                    , sendMessage <|
                        Containers.Message.encodeMsgSend <|
                            Containers.Message.SendAddWord w
                    )

                _ ->
                    ( { model
                        | page =
                            Words
                                { wordsData
                                    | currentWord = Nothing
                                }
                      }
                    , Cmd.none
                    )

        Msg.SendReadyStoryteller ->
            case model.page of
                Page.Started startedData ->
                    ( { model
                        | page = Started { startedData | processState = Started.StorytellerActive }
                      }
                    , sendMessage <|
                        Containers.Message.encodeMsgSend Containers.Message.ReadyStoryteller
                    )

                _ ->
                    ( model, Cmd.none )

        Msg.SendRequestToStart ->
            case model.page of
                Page.Lobby _ ->
                    ( model
                    , sendMessage <|
                        Containers.Message.encodeMsgSend Containers.Message.RequestToStart
                    )

                _ ->
                    ( model, Cmd.none )

        Msg.SendGuessed word ->
            ( model
            , sendMessage <|
                Containers.Message.encodeMsgSend <|
                    Containers.Message.Guess word
            )

        Msg.End ->
            ( { model | page = Home { gameId = Nothing, stats = Nothing } }
            , case model.tokenUser of
                Nothing ->
                    Cmd.none

                Just t ->
                    Home.Http.getStats model t
            )

        Msg.RemoveError ->
            ( { model | error = Nothing }, Cmd.none )

        Msg.GoTo page ->
            case page of
                Page.Home _ ->
                    ( { model | page = page }
                    , case model.tokenUser of
                        Nothing ->
                            Cmd.none

                        Just t ->
                            Home.Http.getStats model t
                    )

                _ ->
                    ( model, Cmd.none )


handleLogin : Model -> Cmd Msg
handleLogin model =
    case model.page of
        Page.Login loginData ->
            case ( loginData.email, loginData.password ) of
                ( Just email, Just password ) ->
                    Login.Http.request model { email = email, password = password }

                _ ->
                    Cmd.none

        _ ->
            Cmd.none


subscriptions : Model -> Sub Msg
subscriptions _ =
    messageReceiver Msg.Rcv


port messageReceiver : (String -> msg) -> Sub msg


port sendJoin : ( String, Int ) -> Cmd msg


port sendHost : ( String, List Int ) -> Cmd msg


port sendMessage : String -> Cmd msg



-- VIEW


view : Model -> Browser.Document Msg
view model =
    case model.page of
        Page.Login loginData ->
            { title = "login", body = Login.View.html model.error loginData }

        Page.Register registerData ->
            { title = "register", body = Register.View.html model.error registerData }

        Page.Change _ changeData ->
            { title = "change", body = Change.View.html model.error changeData }

        Page.Home homeData ->
            { title = "home", body = Home.View.html model.tokenUser model.error homeData }

        Page.Host hostData ->
            { title = "host", body = Host.View.html model.error hostData }

        Page.Lobby lobbyData ->
            { title = "lobby", body = Lobby.View.html lobbyData }

        Page.Words wordsData ->
            { title = "words", body = Words.View.html model.error wordsData }

        Page.Started startedData ->
            { title = "started", body = Started.View.html model.tokenUser model.error startedData }

        Page.Ended players teams ->
            { title = "ended"
            , body =
                case model.tokenUser of
                    Nothing ->
                        Ended.View.html players
                            Nothing
                            teams

                    Just u ->
                        Ended.View.html players (Just u.user) teams
            }


hideError : Cmd Msg
hideError =
    let
        sleep =
            Time.now
                |> Task.andThen
                    (\_ ->
                        Process.sleep
                            (toFloat
                                (3
                                    * 1000
                                )
                            )
                            |> Task.andThen (\() -> Task.succeed ())
                    )
    in
    Task.perform (\() -> Msg.RemoveError) sleep
