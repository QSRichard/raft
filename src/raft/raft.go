package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
const (
	Follower  = 0
	Candidate = 1
	Leader    = 2
)

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	term              int64
	role              int32
	voteFor           int64
	LastLogsIndex     int64
	LastLogsTerm      int64
	LastHeartBeatTime int64
	ElectionTimeout   int64
	CommitIndex       int64
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var isLeader bool
	var term int
	defer rf.mu.Unlock()
	// Your code here (2A).
	rf.mu.Lock()
	if rf.role == Leader {
		isLeader = true
	}
	term = int(rf.term)
	return term, isLeader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

type RequestVoteArgs struct {
	Term         int64
	CandidateId  int64
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteReply struct {
	Term        int64
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int64
	LeaderId     int64
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []byte
	LeaderCommit int64
}

type AppendEntriesReply struct {
	Term    int64
	Success bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.term {
		reply.Term = rf.term
		return
	}

	if rf.role == Leader {
		if args.Term > rf.term {
			rf.term = args.Term
			rf.role = Follower
			rf.voteFor = args.CandidateId
			reply.VoteGranted = true
			reply.Term = args.Term
			return
		} else {
			reply.Term = rf.term
			reply.VoteGranted = false
			return
		}
	}

	if rf.role == Candidate {
		if args.Term > rf.term {
			reply.VoteGranted = true
			reply.Term = args.Term

			rf.term = args.Term
			rf.role = Follower
			rf.voteFor = args.CandidateId
			return
		} else {
			reply.Term = rf.term
			reply.VoteGranted = false
			return
		}
	}

	if rf.role == Follower {
		if rf.voteFor == -1 {
			if rf.term < args.Term && rf.LastLogsIndex <= args.LastLogIndex && rf.LastLogsTerm <= args.LastLogTerm {
				reply.Term = args.Term
				reply.VoteGranted = true

				rf.term = args.Term
				rf.role = Follower
				rf.voteFor = args.CandidateId
			}
		} else {
			if rf.term < args.Term {
				reply.Term = args.Term
				reply.VoteGranted = true

				rf.term = args.Term
				rf.role = Follower
				rf.voteFor = args.CandidateId
			} else {
				reply.Term = rf.term
				reply.VoteGranted = false
			}
		}
	}
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// 自身为Leader,收到了其他Leader发送的AppendEntries
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.role == Leader {
		// 自身term小于入参term
		if rf.term < args.Term {
			// 修改自身状态
			rf.term = args.Term
			rf.role = Follower
			rf.voteFor = args.LeaderId
			rf.LastHeartBeatTime = time.Now().UnixMilli()

			reply.Term = rf.term
			reply.Success = true
		} else {
			reply.Term = rf.term
			reply.Success = false
		}
	} else if rf.role == Follower {
		if rf.term <= args.Term {
			// 修改自身状态
			rf.term = args.Term
			rf.role = Follower
			rf.voteFor = args.LeaderId
			rf.LastHeartBeatTime = time.Now().UnixMilli()

			reply.Term = rf.term
			reply.Success = true
		} else { // 自身term大于入参term
			reply.Term = rf.term
			reply.Success = false
		}
	} else {
		if rf.term <= args.Term {
			// 修改自身状态
			rf.term = args.Term
			rf.role = Follower
			rf.voteFor = args.LeaderId
			rf.LastHeartBeatTime = time.Now().UnixMilli()

			reply.Term = rf.term
			reply.Success = true
		} else { // 自身term大于入参term
			reply.Term = rf.term
			reply.Success = false
		}
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().\
		randomSleepTime := time.Duration(rand.Int63()%rf.ElectionTimeout + rf.ElectionTimeout)
		time.Sleep(time.Millisecond * randomSleepTime)

		// 超时没收到heartBeat, 转换角色
		if rf.role != Leader && time.Now().UnixMilli() > rf.ElectionTimeout+rf.LastHeartBeatTime {
			rf.mu.Lock()
			rf.term += 1
			rf.role = Candidate
			rf.voteFor = int64(rf.me)
			rf.mu.Unlock()

			args := &RequestVoteArgs{
				Term:         rf.term,
				CandidateId:  int64(rf.me),
				LastLogIndex: rf.LastLogsIndex,
				LastLogTerm:  rf.LastLogsTerm,
			}
			wg := sync.WaitGroup{}
			grantCount := atomic.Int64{}
			grantCount.Add(1)
			for index, _ := range rf.peers {
				if index == rf.me {
					continue
				}
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					reply := &RequestVoteReply{}
					rf.sendRequestVote(index, args, reply)
					if reply.VoteGranted {
						grantCount.Add(1)
					} else {
						if reply.Term > rf.term {
							rf.mu.Lock()
							rf.term = reply.Term
							rf.role = Follower
							rf.voteFor = -1
							rf.mu.Unlock()
						}
					}
				}(index)
			}
			wg.Wait()
			if grantCount.Load() > int64(len(rf.peers)/2) {
				rf.mu.Lock()
				rf.role = Leader
				rf.mu.Unlock()
				go rf.heartBeat()
			}
		}
	}
}

func (rf *Raft) heartBeat() {
	for rf.role == Leader {
		args := &AppendEntriesArgs{
			Term:         rf.term,
			LeaderId:     int64(rf.me),
			PrevLogIndex: rf.LastLogsIndex,
			PrevLogTerm:  rf.LastLogsTerm,
			Entries:      []byte{},
			LeaderCommit: rf.CommitIndex,
		}
		wg := sync.WaitGroup{}
		for index, _ := range rf.peers {
			if index == rf.me {
				continue
			}
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				reply := &AppendEntriesReply{}
				rf.sendAppendEntries(index, args, reply)
				if reply.Term > rf.term {
					rf.mu.Lock()
					defer rf.mu.Unlock()
					rf.role = Follower
					rf.term = reply.Term
					rf.voteFor = -1
					rf.LastHeartBeatTime = time.Now().UnixMilli()
				}
			}(index)
		}
		wg.Wait()
		if rf.role != Leader {
			return
		}
		randomSleepTime := time.Duration(rf.ElectionTimeout / 3)
		time.Sleep(time.Millisecond * randomSleepTime)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	// Your initialization code here (2A, 2B, 2C).
	rf.term = 0
	rf.voteFor = -1
	rf.role = Follower
	rf.LastHeartBeatTime = time.Now().UnixMilli()
	rf.ElectionTimeout = 150
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
