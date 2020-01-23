package client

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/go-cid"

	logging "github.com/ipfs/go-log"
	"github.com/ipfs/interface-go-ipfs-core/path"
	pb "github.com/textileio/textile/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var log = logging.Logger("textileclient")

const (
	// chunkSize for store requests.
	chunkSize = 1024
)

// Auth is used to supply the client with authorization credentials.
type Auth struct {
	Token string
	Scope string // user or team ID
}

// Client provides the client api.
type Client struct {
	c    pb.APIClient
	conn *grpc.ClientConn
}

// NewClient starts the client.
func NewClient(target string, creds credentials.TransportCredentials) (*Client, error) {
	var opts []grpc.DialOption
	auth := tokenAuth{}
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
		auth.secure = true
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		c:    pb.NewAPIClient(conn),
		conn: conn,
	}, nil
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Login currently gets or creates a user for the given email address,
// and then waits for email-based verification.
// @todo: Create a dedicated signup flow that collects more info like name, etc.
func (c *Client) Login(ctx context.Context, email string) (*pb.LoginReply, error) {
	return c.c.Login(ctx, &pb.LoginRequest{Email: email})
}

// Switch changes session scope.
func (c *Client) Switch(ctx context.Context, auth Auth) error {
	_, err := c.c.Switch(authCtx(ctx, auth), &pb.SwitchRequest{})
	return err
}

// Logout deletes a remote session.
func (c *Client) Logout(ctx context.Context, auth Auth) error {
	_, err := c.c.Logout(authCtx(ctx, auth), &pb.LogoutRequest{})
	return err
}

// Whoami returns session info.
func (c *Client) Whoami(ctx context.Context, auth Auth) (*pb.WhoamiReply, error) {
	return c.c.Whoami(authCtx(ctx, auth), &pb.WhoamiRequest{})
}

// AddTeam add a new team.
func (c *Client) AddTeam(ctx context.Context, name string, auth Auth) (*pb.AddTeamReply, error) {
	return c.c.AddTeam(authCtx(ctx, auth), &pb.AddTeamRequest{Name: name})
}

// GetTeam returns a team by ID.
func (c *Client) GetTeam(ctx context.Context, teamID string, auth Auth) (*pb.GetTeamReply, error) {
	return c.c.GetTeam(authCtx(ctx, auth), &pb.GetTeamRequest{
		ID: teamID,
	})
}

// ListTeams returns a list of authorized teams.
func (c *Client) ListTeams(ctx context.Context, auth Auth) (*pb.ListTeamsReply, error) {
	return c.c.ListTeams(authCtx(ctx, auth), &pb.ListTeamsRequest{})
}

// RemoveTeam removes a team by ID.
func (c *Client) RemoveTeam(ctx context.Context, teamID string, auth Auth) error {
	_, err := c.c.RemoveTeam(authCtx(ctx, auth), &pb.RemoveTeamRequest{
		ID: teamID,
	})
	return err
}

// InviteToTeam invites the given email to a team by ID.
func (c *Client) InviteToTeam(ctx context.Context, teamID, email string, auth Auth) (*pb.InviteToTeamReply, error) {
	return c.c.InviteToTeam(authCtx(ctx, auth), &pb.InviteToTeamRequest{
		ID:    teamID,
		Email: email,
	})
}

// LeaveTeam removes the authorized user from a team by ID.
func (c *Client) LeaveTeam(ctx context.Context, teamID string, auth Auth) error {
	_, err := c.c.LeaveTeam(authCtx(ctx, auth), &pb.LeaveTeamRequest{
		ID: teamID,
	})
	return err
}

// AddProject add a new project under the given scope.
func (c *Client) AddProject(ctx context.Context, name string, auth Auth) (*pb.AddProjectReply, error) {
	return c.c.AddProject(authCtx(ctx, auth), &pb.AddProjectRequest{
		Name: name,
	})
}

// GetProject returns a project by ID.
func (c *Client) GetProject(ctx context.Context, projID string, auth Auth) (*pb.GetProjectReply, error) {
	return c.c.GetProject(authCtx(ctx, auth), &pb.GetProjectRequest{
		ID: projID,
	})
}

// ListProjects returns a list of all authorized projects.
func (c *Client) ListProjects(ctx context.Context, auth Auth) (*pb.ListProjectsReply, error) {
	return c.c.ListProjects(authCtx(ctx, auth), &pb.ListProjectsRequest{})
}

// RemoveProject removes a project by ID.
func (c *Client) RemoveProject(ctx context.Context, projID string, auth Auth) error {
	_, err := c.c.RemoveProject(authCtx(ctx, auth), &pb.RemoveProjectRequest{
		ID: projID,
	})
	return err
}

// AddAppToken add a new app token under the given project.
func (c *Client) AddAppToken(ctx context.Context, projID string, auth Auth) (*pb.AddAppTokenReply, error) {
	return c.c.AddAppToken(authCtx(ctx, auth), &pb.AddAppTokenRequest{
		ProjectID: projID,
	})
}

// ListAppTokens returns a list of all app tokens for the given project.
func (c *Client) ListAppTokens(ctx context.Context, projID string, auth Auth) (*pb.ListAppTokensReply, error) {
	return c.c.ListAppTokens(authCtx(ctx, auth), &pb.ListAppTokensRequest{
		ProjectID: projID,
	})
}

// RemoveAppToken removes an app token by ID.
func (c *Client) RemoveAppToken(ctx context.Context, tokenID string, auth Auth) error {
	_, err := c.c.RemoveAppToken(authCtx(ctx, auth), &pb.RemoveAppTokenRequest{
		ID: tokenID,
	})
	return err
}

type storeResult struct {
	path path.Resolved
	err  error
}

// Store sends a file.
func (c *Client) Store(ctx context.Context, projID string, filePath string, progress chan int64, auth Auth) (path.Resolved, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stream, err := c.c.Store(authCtx(ctx, auth))
	if err != nil {
		return nil, err
	}
	if err = stream.Send(&pb.StoreRequest{
		Payload: &pb.StoreRequest_ProjectID{
			ProjectID: projID,
		},
	}); err != nil {
		return nil, err
	}

	waitCh := make(chan storeResult)
	go func() {
		defer close(waitCh)
		for {
			rep, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				waitCh <- storeResult{err: err}
				return
			}
			switch payload := rep.GetPayload().(type) {
			case *pb.StoreReply_Event_:
				if payload.Event.Path != "" {
					id, err := cid.Parse(payload.Event.Path)
					if err != nil {
						waitCh <- storeResult{err: err}
						return
					}
					waitCh <- storeResult{path: path.IpfsPath(id)}
				} else {
					progress <- payload.Event.Bytes
				}
			case *pb.StoreReply_Error:
				waitCh <- storeResult{err: fmt.Errorf(payload.Error)}
				return
			default:
				waitCh <- storeResult{err: fmt.Errorf("invalid reply")}
				return
			}
		}
	}()

	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if err = stream.Send(&pb.StoreRequest{
			Payload: &pb.StoreRequest_Chunk{
				Chunk: buf[:n],
			},
		}); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	if err = stream.CloseSend(); err != nil {
		return nil, err
	}
	res := <-waitCh
	return res.path, res.err
}

type authKey string

func authCtx(ctx context.Context, auth Auth) context.Context {
	ctx = context.WithValue(ctx, authKey("token"), auth.Token)
	if auth.Scope != "" {
		ctx = context.WithValue(ctx, authKey("scope"), auth.Scope)
	}
	return ctx
}

type tokenAuth struct {
	secure bool
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	token, ok := ctx.Value(authKey("token")).(string)
	if ok && token != "" {
		md["Authorization"] = "Bearer " + token
	}
	scope, ok := ctx.Value(authKey("scope")).(string)
	if ok && scope != "" {
		md["X-Scope"] = scope
	}
	return md, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return t.secure
}
