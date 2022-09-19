/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package network

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fabric/controller/command"
	"github.com/openziti/fabric/controller/db"
	"github.com/openziti/fabric/controller/fields"
	"github.com/openziti/fabric/controller/idgen"
	"github.com/openziti/fabric/controller/models"
	"github.com/openziti/fabric/ioc"
	"github.com/openziti/fabric/pb/cmd_pb"
	//"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

const (
	CreateDecoder = "CreateDecoder"
	UpdateDecoder = "UpdateDecoder"
	DeleteDecoder = "DeleteDecoder"
)

type Managers struct {
	network     *Network
	db          boltz.Db
	stores      *db.Stores
	Terminators *TerminatorManager
	Routers     *RouterManager
	Services    *ServiceManager
	Inspections *InspectionsManager
	Command     *CommandManager
	Dispatcher  command.Dispatcher
	Registry    ioc.Registry
}

func (self *Managers) getDb() boltz.Db {
	return self.db
}

func (self *Managers) Dispatch(command command.Command) error {
	return self.Dispatcher.Dispatch(command)
}

type creator[T models.Entity] interface {
	command.EntityCreator[T]
	Dispatch(cmd command.Command) error
}

type updater[T models.Entity] interface {
	command.EntityUpdater[T]
	Dispatch(cmd command.Command) error
}

func DispatchCreate[T models.Entity](c creator[T], entity T) error {
	if entity.GetId() == "" {
		id, err := idgen.NewUUIDString()
		if err != nil {
			return err
		}
		fmt.Println("ID CREATED?", id)
		fmt.Println("OF THIS ENTITY?", entity)
		fmt.Println("THE TAGS?", entity.GetTags())
		entity.SetId(id)
	}

	cmd := &command.CreateEntityCommand[T]{
		Creator: c,
		Entity:  entity,
	}

	fmt.Println("THE ACTUAL COMMAND?-----------------------------", cmd.Entity)

	return c.Dispatch(cmd)
}

func DispatchUpdate[T models.Entity](u updater[T], entity T, updatedFields fields.UpdatedFields) error {
	cmd := &command.UpdateEntityCommand[T]{
		Updater:       u,
		Entity:        entity,
		UpdatedFields: updatedFields,
	}

	return u.Dispatch(cmd)
}

type createDecoderF func(cmd *cmd_pb.CreateEntityCommand) (command.Command, error)

func RegisterCreateDecoder[T models.Entity](managers *Managers, creator command.EntityCreator[T]) {
	entityType := creator.GetEntityTypeId()
	managers.Registry.RegisterSingleton(entityType+CreateDecoder, createDecoderF(func(cmd *cmd_pb.CreateEntityCommand) (command.Command, error) {
		entity, err := creator.Unmarshall(cmd.EntityData)
		if err != nil {
			return nil, err
		}
		return &command.CreateEntityCommand[T]{
			Entity:  entity,
			Creator: creator,
			Flags:   cmd.Flags,
		}, nil
	}))
}

type updateDecoderF func(cmd *cmd_pb.UpdateEntityCommand) (command.Command, error)

func RegisterUpdateDecoder[T models.Entity](managers *Managers, updater command.EntityUpdater[T]) {
	entityType := updater.GetEntityTypeId()
	managers.Registry.RegisterSingleton(entityType+UpdateDecoder, updateDecoderF(func(cmd *cmd_pb.UpdateEntityCommand) (command.Command, error) {
		entity, err := updater.Unmarshall(cmd.EntityData)
		if err != nil {
			return nil, err
		}
		return &command.UpdateEntityCommand[T]{
			Entity:        entity,
			Updater:       updater,
			UpdatedFields: fields.SliceToUpdatedFields(cmd.UpdatedFields),
			Flags:         cmd.Flags,
		}, nil
	}))
}

type deleteDecoderF func(cmd *cmd_pb.DeleteEntityCommand) (command.Command, error)

func RegisterDeleteDecoder(managers *Managers, deleter command.EntityDeleter) {
	entityType := deleter.GetEntityTypeId()
	managers.Registry.RegisterSingleton(entityType+DeleteDecoder, deleteDecoderF(func(cmd *cmd_pb.DeleteEntityCommand) (command.Command, error) {
		return &command.DeleteEntityCommand{
			Deleter: deleter,
			Id:      cmd.EntityId,
		}, nil
	}))
}

func RegisterManagerDecoder[T models.Entity](managers *Managers, ctrl command.EntityManager[T]) {
	RegisterCreateDecoder[T](managers, ctrl)
	RegisterUpdateDecoder[T](managers, ctrl)
	RegisterDeleteDecoder(managers, ctrl)
}

func NewManagers(network *Network, dispatcher command.Dispatcher, db boltz.Db, stores *db.Stores) *Managers {
	result := &Managers{
		network:    network,
		db:         db,
		stores:     stores,
		Dispatcher: dispatcher,
		Registry:   ioc.NewRegistry(),
	}
	result.Command = newCommandManager(result)
	result.Terminators = newTerminatorManager(result)
	result.Routers = newRouterManager(result)
	result.Services = newServiceManager(result)
	result.Inspections = NewInspectionsManager(network)
	if result.Dispatcher == nil {
		//devVersion := versions.MustParseSemVer("0.0.0")
		//version := versions.MustParseSemVer(network.VersionProvider.Version())
		result.Dispatcher = &command.LocalDispatcher{
			EncodeDecodeCommands: false, //change later to see why
			//EncodeDecodeCommands: devVersion.Equals(version),
		}
	}
	result.Command.registerGenericCommands()

	RegisterManagerDecoder[*Service](result, result.Services)
	RegisterManagerDecoder[*Router](result, result.Routers)
	RegisterManagerDecoder[*Terminator](result, result.Terminators)

	return result
}

type Controller[T models.Entity] interface {
	models.EntityRetriever[T]
	getManagers() *Managers
}

func newBaseEntityManager[T models.Entity](managers *Managers, store boltz.CrudStore, newModelEntity func() T) baseEntityManager[T] {
	return baseEntityManager[T]{
		BaseEntityManager: models.BaseEntityManager{
			Store: store,
		},
		Managers:       managers,
		newModelEntity: newModelEntity,
	}
}

type baseEntityManager[T models.Entity] struct {
	models.BaseEntityManager
	*Managers
	newModelEntity func() T
	populateEntity func(entity T, tx *bbolt.Tx, boltEntity boltz.Entity) error
}

func (self *baseEntityManager[T]) GetEntityTypeId() string {
	// default this to the store entity type and let individual managers override it where
	// needed to avoid collisions (e.g. edge service/router)
	return self.GetStore().GetEntityType()
}

func (self *baseEntityManager[T]) Delete(id string) error {
	cmd := &command.DeleteEntityCommand{
		Deleter: self,
		Id:      id,
	}
	return self.Managers.Dispatch(cmd)
}

func (self *baseEntityManager[T]) ApplyDelete(cmd *command.DeleteEntityCommand) error {
	return self.db.Update(func(tx *bbolt.Tx) error {
		ctx := boltz.NewMutateContext(tx)
		return self.Store.DeleteById(ctx, cmd.Id)
	})
}

func (ctrl *baseEntityManager[T]) BaseLoad(id string) (T, error) {
	entity := ctrl.newModelEntity()
	if err := ctrl.readEntity(id, entity); err != nil {
		var defaultValue T
		return defaultValue, err
	}
	return entity, nil
}

func (ctrl *baseEntityManager[T]) BaseLoadInTx(tx *bbolt.Tx, id string) (T, error) {
	entity := ctrl.newModelEntity()
	if err := ctrl.readEntityInTx(tx, id, entity); err != nil {
		var defaultValue T
		return defaultValue, err
	}
	return entity, nil
}

func (ctrl *baseEntityManager[T]) getManagers() *Managers {
	return ctrl.Managers
}

func (ctrl *baseEntityManager[T]) readEntity(id string, modelEntity T) error {
	return ctrl.db.View(func(tx *bbolt.Tx) error {
		return ctrl.readEntityInTx(tx, id, modelEntity)
	})
}

func (ctrl *baseEntityManager[T]) readEntityInTx(tx *bbolt.Tx, id string, modelEntity T) error {
	boltEntity := ctrl.GetStore().NewStoreEntity()
	found, err := ctrl.GetStore().BaseLoadOneById(tx, id, boltEntity)
	if err != nil {
		return err
	}
	if !found {
		return boltz.NewNotFoundError(ctrl.GetStore().GetSingularEntityType(), "id", id)
	}

	return ctrl.populateEntity(modelEntity, tx, boltEntity)
}

func (ctrl *baseEntityManager[T]) BaseList(query string) (*models.EntityListResult[T], error) {
	result := &models.EntityListResult[T]{Loader: ctrl}
	err := ctrl.ListWithHandler(query, result.Collect)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ctrl *baseEntityManager[T]) ListWithHandler(queryString string, resultHandler models.ListResultHandler) error {
	return ctrl.db.View(func(tx *bbolt.Tx) error {
		return ctrl.ListWithTx(tx, queryString, resultHandler)
	})
}

func (ctrl *baseEntityManager[T]) BasePreparedList(query ast.Query) (*models.EntityListResult[T], error) {
	result := &models.EntityListResult[T]{Loader: ctrl}
	err := ctrl.PreparedListWithHandler(query, result.Collect)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ctrl *baseEntityManager[T]) PreparedListWithHandler(query ast.Query, resultHandler models.ListResultHandler) error {
	return ctrl.db.View(func(tx *bbolt.Tx) error {
		return ctrl.PreparedListWithTx(tx, query, resultHandler)
	})
}

func (ctrl *baseEntityManager[T]) PreparedListAssociatedWithHandler(id string, association string, query ast.Query, handler models.ListResultHandler) error {
	return ctrl.db.View(func(tx *bbolt.Tx) error {
		return ctrl.PreparedListAssociatedWithTx(tx, id, association, query, handler)
	})
}

type boltEntitySource interface {
	models.Entity
	toBolt() boltz.Entity
}

func (ctrl *baseEntityManager[T]) updateGeneral(modelEntity boltEntitySource, checker boltz.FieldChecker) error {
	return ctrl.db.Update(func(tx *bbolt.Tx) error {
		ctx := boltz.NewMutateContext(tx)
		existing := ctrl.GetStore().NewStoreEntity()
		found, err := ctrl.GetStore().BaseLoadOneById(tx, modelEntity.GetId(), existing)
		if err != nil {
			return err
		}
		if !found {
			return boltz.NewNotFoundError(ctrl.GetStore().GetSingularEntityType(), "id", modelEntity.GetId())
		}

		boltEntity := modelEntity.toBolt()

		if err := ctrl.ValidateNameOnUpdate(ctx, boltEntity, existing, checker); err != nil {
			return err
		}

		if err := ctrl.GetStore().Update(ctx, boltEntity, checker); err != nil {
			pfxlog.Logger().WithError(err).Errorf("could not update %v entity", ctrl.GetStore().GetEntityType())
			return err
		}
		return nil
	})
}
