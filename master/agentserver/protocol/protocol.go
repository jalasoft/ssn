package protocol

import (
	"fmt"
	"log"
	"time"

	"github.com/jalasoft/ssn/master/registry"
)

func HelloState(session Session, c <-chan interface{}) {

	readChan := make(chan []byte)
	go session.Read(readChan)

	select {
	case <-c:
		log.Printf("Closing session")
		go ByeState(session)
		break

	case msg, ok := <-readChan:
		if !ok {
			session.LogInfo("Reading failed, closing")
			session.Close()
			break
		}

		helloMsg, err := ParseHelloMessage(msg)
		if err != nil {
			session.LogError(err)
			go ByeState(session)
			return
		}
		session.LogInfo(fmt.Sprintf("Obtained HELLO %s: %v", helloMsg.Name, helloMsg.Props))
		session.LogInfo("sending [HI]")
		if session.Write("[HI]") {
			agent := &registry.Agent{
				Id:     session.Id(),
				Name:   helloMsg.Name,
				Traits: helloMsg.Props,
				Skills: make([]registry.Skill, 0),
			}
			go TellMeSkillsState(session, agent, c)
		} else {
			session.LogInfo("Sending [HI] failed.")
		}
		break
	}
}

func TellMeSkillsState(session Session, agent *registry.Agent, c <-chan interface{}) {

	if !session.Write("[TELLMESKILLS]") {
		go ByeState(session)
		return
	}

	readChan := make(chan []byte)
	go session.Read(readChan)

	select {
	case <-c:
		session.LogInfo("Closing session")
		go ByeState(session)
		break

	case msg, ok := <-readChan:
		if !ok {
			session.LogInfo("Reading failed, closing")
			session.Close()
			break
		}

		skillMsg, err := ParseSkillMessage(msg)

		if err != nil {
			session.LogError(err)
			go ByeState(session)
			break
		}

		session.LogInfo(fmt.Sprintf("Obtained message SKILL: %v", skillMsg))

		skill := registry.Skill{
			Name:   skillMsg.Name,
			Type:   skillMsg.Type,
			Traits: skillMsg.Props,
		}
		agent.Skills = append(agent.Skills, skill)
		go TellMeSkillsContinuedState(session, agent, c)
	}
}

func TellMeSkillsContinuedState(session Session, agent *registry.Agent, c <-chan interface{}) {

	readChan := make(chan []byte)
	go session.Read(readChan)

	select {
	case <-c:
		session.LogInfo("Closing session")
		go ByeState(session)
		break

	case msgBytes, ok := <-readChan:
		if !ok {
			session.LogInfo("Reading failed, closing")
			session.Close()
			break
		}
		if IsThatsAllsMessage(msgBytes) {
			session.LogInfo("All skills obtained")
			session.Write("[THANKS]")
			registry.Register(agent)
			go IAmHereState(session, c)
			break
		}

		skillMsg, err := ParseSkillMessage(msgBytes)

		if err != nil {
			session.LogError(err)
			go ByeState(session)
			break
		}

		session.LogInfo(fmt.Sprintf("Obtained message SKILL: %v", skillMsg))
		skill := registry.Skill{
			Name:   skillMsg.Name,
			Type:   skillMsg.Type,
			Traits: skillMsg.Props,
		}
		agent.Skills = append(agent.Skills, skill)
		go TellMeSkillsContinuedState(session, agent, c)
	}
}

func IAmHereState(session Session, c <-chan interface{}) {

root:
	for {
		readChan := make(chan []byte)
		go session.Read(readChan)

		timer := time.NewTimer(10 * time.Second)

		select {
		case <-c:
			session.LogInfo("Closing session")
			go ByeState(session)
			break

		case msgBytes, ok := <-readChan:
			if !ok {
				session.LogInfo("Reading failed, closing")
				registry.Unregister(session.Id())
				session.Close()
				break root
			}
			if IsIamStillHereMessage(msgBytes) {
				session.LogInfo("[IAMSTILLHERE] received")
				break
			}

			if IsByeMessage(msgBytes) {
				session.LogInfo("[BYE] received")
				registry.Unregister(session.Id())
				session.Close()
				break root
			}

			session.LogError(fmt.Errorf("Unexpected message obtained: '%s'", string(msgBytes)))
			go ByeState(session)
			break root

		case <-timer.C:
			session.LogInfo("Session timeouted")
			go ByeState(session)
			break root
		}
	}
}

func ByeState(session Session) {
	registry.Unregister(session.Id())
	session.LogInfo("sending [BYE]")
	session.Write("[BYE]")
	session.Close()
}
