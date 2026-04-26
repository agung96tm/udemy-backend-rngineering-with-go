package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"socialv2/internal/store"
	"strings"
)

var usernames = []string{"tiago", "bob", "radios", "torin", "charlie", "bolin", "radian", "alex", "mila", "niko", "luna", "jax", "sophia", "leo", "aria", "kai", "elya", "zane", "nova", "finn", "clara", "ryan", "ivy", "ezra", "lyra", "omar", "vivi", "tobias", "nina", "kaiya", "silas", "celeste", "maverick", "seren", "jaxon", "freya", "damian", "elara", "rowan", "lila", "orion", "skye", "alina", "evan", "ariah", "lucas", "zoe", "adrian", "amelia", "talia", "nash", "elyse", "phoenix", "rory", "caden", "selene", "matteo", "ophelia", "kairos", "lyric", "ashton", "bella", "cyrus", "dahlia", "emrys", "fiona", "grayson", "hazel", "isaac", "juno", "kian", "livia", "marcus", "nadia", "oliver", "penelope", "quinn", "ryker", "seraphina", "thomas", "uma", "vincent", "willow", "xander", "yasmin", "zayden", "amelie", "bruno", "cassandra", "dante", "elise", "fabian", "gabriella", "hugo", "isla", "jaxie", "kairo", "leona", "micah", "nova", "orla", "paxton", "quincy", "rowena", "silvan", "tristan", "ursula", "valen", "wren", "xena", "yara", "zephyr", "adeline", "beau", "celina", "dorian", "elara", "felix", "gianna", "holden", "imogen", "jace", "kaia", "lucian", "mira", "niko", "octavia", "pax", "quorra", "rylan", "selah", "thaddeus", "ulric", "vivian", "weston", "xavia", "yasuo", "zinnia", "alessio", "bianca", "caius", "dahl", "emilia", "fabio", "gabriel", "hester", "ignatius", "jade", "klaus", "lorelei", "maia", "nathaniel", "odette", "pierre", "quinnley", "rebecca", "soren", "tessa", "uliana", "valeria", "winston", "xavier", "yasmina", "zev", "anastasia", "blake", "cassius", "daphne", "emmett", "fleur", "giovanni", "harper", "ion", "jessica", "kian", "lola", "matteo", "naya", "oscar", "phoebe", "quintus", "raegan", "sylas", "tobias", "ursula", "victor", "willa", "xander", "yvette", "zayden", "amelia", "bastian", "clara", "dante", "elodie", "fabian", "genevieve", "hugo", "isabel", "jaxon", "kaia", "leo", "marina", "nico", "olivia", "peter", "quinn", "ryan", "selene", "thomas", "ulysses", "violet", "william", "xenia", "yasmin", "zane", "alina", "bruno", "carmen", "damian", "elara", "fabio", "gianna", "harrison", "imogen", "jax", "kairos", "livia", "micah", "nina", "orion", "paul", "quorra", "rowan", "silas", "tobias", "ulyana", "valen", "wesley", "xara", "yara", "zephyr", "adrian", "bella", "cassian", "dahlia", "emrys", "fiona", "gideon", "hazel", "isaac", "juno", "kieran", "lorelei", "marcus", "nadia", "oliver", "penelope", "quincy", "ryker", "seraphina", "thaddeus", "uma", "vincent", "willow", "xander", "yasmin", "zayden", "amelie", "bruno", "cassandra", "dante", "elise", "fabian", "gabriella", "hugo", "isla", "jaxie", "kairo", "leona", "micah", "nova", "orla", "paxton", "quinnley", "rowena", "silvan", "tristan", "ursula", "valeria", "wren", "xena", "yara", "zephyr", "adeline", "beau", "celina", "dorian", "elara", "felix", "gianna", "holden", "imogen", "jace", "kaia", "lucian", "mira", "niko", "octavia", "pax", "quorra", "rylan", "selah", "thaddeus", "ulric", "vivian", "weston", "xavia", "yasuo", "zinnia"}
var titles = []string{
	"Exploring the Future of AI",
	"Building Scalable Web Apps",
	"The Power of Cloud Computing",
	"Mastering Data Structures",
	"Understanding Blockchain",
	"Introduction to Machine Learning",
	"Optimizing Database Performance",
	"Design Patterns in Go",
	"Securing Your Web Applications",
	"Building RESTful APIs",
	"Microservices Architecture",
	"Getting Started with Kubernetes",
	"Advanced SQL Techniques",
	"Deploying Apps with Docker",
	"Concurrency in Modern Programming",
	"Testing Strategies for Developers",
	"Improving Frontend Performance",
	"Version Control Best Practices",
	"Clean Code Principles",
	"Automating Workflows with CI/CD",
	"Scaling Teams in Tech Startups",
	"Design Thinking for Developers",
	"How to Write Maintainable Code",
	"The Rise of Serverless Computing",
	"Modern DevOps Practices",
	"Handling Big Data Efficiently",
	"API Security Fundamentals",
	"Debugging Complex Applications",
	"Designing for Accessibility",
	"From Monolith to Microservices",
	"Understanding Software Licensing",
	"Best Practices for Code Reviews",
	"Improving Developer Productivity",
	"Optimizing Go Routines",
	"Introduction to Event-Driven Systems",
	"Containerization Made Easy",
	"Integrating AI into Business Apps",
	"Improving UX with Data Insights",
	"The Art of Writing Documentation",
	"Managing Technical Debt",
	"Testing REST APIs with Postman",
	"Improving System Observability",
	"Logging and Monitoring Strategies",
	"Getting Started with GraphQL",
	"Effective Error Handling in Go",
	"Introduction to TypeScript",
	"Securing APIs with OAuth2",
	"Database Indexing Explained",
	"Understanding Load Balancers",
	"Handling Concurrency Safely",
	"Effective Caching Strategies",
	"Event Sourcing for Beginners",
	"Introduction to Domain-Driven Design",
	"Optimizing Docker Images",
	"Improving Query Performance",
	"Working with Message Queues",
	"Designing Resilient Systems",
	"Modern API Design Principles",
	"Automated Testing in CI Pipelines",
	"Scalable Logging with ELK Stack",
	"Improving Team Collaboration",
}
var tags = []string{
	"go", "golang", "backend", "frontend", "javascript", "typescript", "python",
	"machine-learning", "ai", "docker", "kubernetes", "microservices", "graphql",
	"devops", "sql", "nosql", "database", "rest-api", "security", "testing",
	"clean-code", "design-patterns", "cloud", "aws", "gcp", "azure", "linux",
	"serverless", "ci-cd", "performance", "optimization", "data-engineering",
	"react", "vue", "nextjs", "fastapi", "nestjs", "express", "go-routines",
}

func getString(list []string, count int) string {
	if len(list) == 0 || count <= 0 {
		return ""
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = list[rand.Intn(len(list))]
	}
	return strings.Join(result, " ")
}

func Seed(store store.Storage, db *sql.DB) error {
	ctx := context.Background()
	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			tx.Rollback()
			log.Println("Error creating user", user, err)
			return nil
		}
	}
	tx.Commit()

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post", post, err)
			return nil
		}
	}

	comments := generateComments(10, users, posts)
	for _, comment := range comments {
		if err := store.Comment.Create(ctx, comment); err != nil {
			log.Println("Error creating comment", comment, err)
			return nil
		}
	}

	log.Println("Seeding Completed")
	return nil
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)
	for i := 0; i < num; i++ {
		name := usernames[rand.Intn(len(usernames))]
		users[i] = &store.User{
			Name:  fmt.Sprintf("%s%d", name, i),
			Email: fmt.Sprintf("%s%d@example.com", name, i),
		}
	}
	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   getString(titles, 2),
			Content: getString(titles, 20),
			Tags:    strings.Split(getString(tags, 3), " "),
		}
	}
	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]

		comments[i] = &store.Comment{
			UserID:  user.ID,
			PostID:  post.ID,
			Content: getString(titles, 20),
		}
	}
	return comments
}
