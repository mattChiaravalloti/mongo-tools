package mongoplay

import (
	mgo "github.com/10gen/llmgo"
	"github.com/10gen/llmgo/bson"
	"testing"
)

//TestCommandsAgainstAuthedDBWhenAuthed tests some basic commands against a database that requires authenticaiton
//when the driver has proper authentication credentials
func TestCommandsAgainstAuthedDBWhenAuthed(t *testing.T) {
	if !authTestServerMode {
		t.Skipf("Skipping auth test with non-auth DB")
	}
	if err := teardownDB(); err != nil {
		t.Error(err)
	}
	numInserts := 20
	generator := newRecordedOpGenerator()

	go func() {
		defer close(generator.opChan)
		err := generator.generateInsertHelper("Authed Insert Test", 0, numInserts)
		if err != nil {
			t.Error(err)
		}
	}()
	statColl := NewBufferedStatCollector()
	context := NewExecutionContext(statColl)
	err := Play(context, generator.opChan, testSpeed, authTestServerUrl, 1, 10)
	if err != nil {
		t.Error(err)
	}

	session, err := mgo.Dial(authTestServerUrl)
	coll := session.DB(testDB).C(testCollection)

	iter := coll.Find(bson.D{}).Sort("docNum").Iter()
	ind := 0
	result := testDoc{}

	for iter.Next(&result) {
		if err := iter.Err(); err != nil {
			t.Errorf("Iterator returned an error: %v\n", err)
		}
		if result.DocumentNumber != ind {
			t.Errorf("Inserted document number did not match expected document number. Found: %v -- Expected: %v", result.DocumentNumber, ind)
		}
		if result.Name != "Authed insert test" {
			t.Errorf("Inserted document name did not match expected name. Found %v -- Expected: %v", result.Name, "Authed insert test")
		}
		if !result.Success {
			t.Errorf("Inserted document field 'Success' was expected to be true, but was false")
		}
		ind++
	}
	if err := iter.Close(); err != nil {
		t.Error(err)
	}
	if err := teardownDB(); err != nil {
		t.Error(err)
	}

}

//TestCommandsAgainstAuthedDBWhenNotAuthed tests some basic commands against a database that requires authentication
//when the driver does not have proper authenticaiton. It generates a series of inserts and ensures that the docs they are attempting
//to insert are not later found in the database
func TestCommandsAgainstAuthedDBWhenNotAuthed(t *testing.T) {
	if !authTestServerMode {
		t.Skipf("Skipping auth test with non-auth DB")
	}
	if err := teardownDB(); err != nil {
		t.Error(err)
	}
	numInserts := 3
	generator := newRecordedOpGenerator()

	go func() {
		defer close(generator.opChan)
		err := generator.generateInsertHelper("Non-Authed Insert Test", 0, numInserts)
		if err != nil {
			t.Error(err)
		}
	}()
	statColl := NewBufferedStatCollector()
	context := NewExecutionContext(statColl)
	err := Play(context, generator.opChan, testSpeed, nonAuthTestServerUrl, 1, 10)
	if err != nil {
		t.Error(err)
	}
	session, err := mgo.Dial(authTestServerUrl)
	coll := session.DB(testDB).C(testCollection)
	num, err := coll.Find(bson.D{}).Count()
	if err != nil {
		t.Error(err)
	}
	if num != 0 {
		t.Errorf("Collection contained documents, expected it to be empty. Num: %d\n", num)
	}
	if err := teardownDB(); err != nil {
		t.Error(err)
	}

}
