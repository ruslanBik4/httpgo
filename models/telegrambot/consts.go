// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telegrambot

import "github.com/pkg/errors"

// request params for
const (
	baseURL       string = "https://api.telegram.org/bot"
	maxMessLength int    = 4050
)

// environment values for init telegramBot
const (
	TB_TOKEN = "TBTOKEN"
	CHAT_ID  = "TBCHATID"
)

const (
	cmdGetMe                     = "getMe"
	cmdSendMes                   = "sendMessage"
	cmdFormatOpt                 = "Formatting options"
	cmdFwdMess                   = "forwardMessage"
	cmdSendPhoto                 = "sendPhoto"
	cmdSendAudio                 = "sendAudio"
	cmdSendDocument              = "sendDocument"
	cmdSendVideo                 = "sendVideo"
	cmdSendAnimation             = "sendAnimation"
	cmdSendVoice                 = "sendVoice"
	cmdSendVideoNote             = "sendVideoNote"
	cmdSendMediaGroup            = "sendMediaGroup"
	cmdSendLocation              = "sendLocation"
	cmdEditMesLiveLoc            = "editMessageLiveLocation"
	cmdStopMesLiveLoc            = "stopMessageLiveLocation"
	cmdSendVenue                 = "sendVenue"
	cmdSendContact               = "sendContact"
	cmdSendPoll                  = "sendPoll"
	cmdSendChatAction            = "sendChatAction"
	cmdGetUsrPflPhoto            = "getUserProfilePhotos"
	cmdGetFile                   = "getFile"
	cmdKickChMbr                 = "kickChatMember"
	cmdUnbanChMbr                = "unbanChatMember"
	cmdResctrictChMbr            = "restrictChatMember"
	cmdPromoteChMbr              = "promoteChatMember"
	cmdSetChPerm                 = "setChatPermissions"
	cmdExportChLink              = "exportChatInviteLink"
	cmdSetChPhoto                = "setChatPhoto"
	cmdDelChPhoto                = "deleteChatPhoto"
	cmdSetChTitle                = "setChatTitle"
	cmdSetChDesc                 = "setChatDescription"
	cmdPinChMes                  = "pinChatMessage"
	cmdUnpinCHMes                = "unpinChatMessage"
	cmdLeaveCH                   = "leaveChat"
	cmdGetChat                   = "getChat"
	cmdSendgetChatAdministrators = "getChatAdministrators"
	cmdGetChMbrsCount            = "getChatMembersCount"
	cmdGetChMbr                  = "getChatMember"
	cmdSetChatStickerSet         = "setChatStickerSet"
	cmddeleteChatStickerSet      = "deleteChatStickerSet"
	cmdanswerCallbackQuery       = "answerCallbackQuery"
	cmdInlineMThd                = "Inline mode methods"
	cmdgetUpdates                = "getUpdates"
)

const maxStack = 30

const errMessBadBotParams = "TelegramBot struct param missing"

type ErrBadBotParams struct {
	BadParam string
}

func (err ErrBadBotParams) Error() string {
	return errMessBadBotParams + ": " + err.BadParam
}

var (
	ErrBadTelegramBot         = errors.New("Bad TelegramBot parameters")
	ErrTelegramBotMultiple500 = errors.New("Telegram does not response, multiple 500")
	ErrEmptyMessText          = errors.New("Telegram message is empty")
	ErrTooLongMessText        = errors.New("Telegram message is too long")
)
